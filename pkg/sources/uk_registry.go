package sources

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gghez/gopi/internal/html"
	"github.com/gghez/gopi/internal/types"
	"github.com/gghez/gopi/pkg/search"
)

const websiteRootURL string = "https://find-and-update.company-information.service.gov.uk"

type UKCompany struct {
	ID                 string
	Name               string
	Role               string
	Status             string
	Address            string
	RoleAppointedOn    types.Date
	ResignedOn         *types.Date
	Nationality        *string
	CountryOfResidence *string
	Occupation         *string
	URL                string
}

type UKSearchResult struct {
	ID          string
	Name        string
	URL         string
	DateOfBirth *types.MonthDate
	Companies   []UKCompany
}

type ukRegistrySearchEngine struct {
}

const sourceName = "uk_registry"

var rxOfficerID *regexp.Regexp = regexp.MustCompile(`/officers/(?P<OfficerID>\w+)/appointments`)

func NewUKRegistrySearchEngine() search.Searcher[*UKSearchResult] {
	return new(ukRegistrySearchEngine)
}

func (engine *ukRegistrySearchEngine) Source() string {
	return sourceName
}

func (engine *ukRegistrySearchEngine) Search(query string) ([]*UKSearchResult, error) {
	log.Info().Msgf("search uk companies registry: %q", query)
	urlEncodedQuery := url.QueryEscape(query)

	url := fmt.Sprintf("%s/search/officers?q=%s", websiteRootURL, urlEncodedQuery)
	root, err := html.Parse(url)
	if err != nil {
		return nil, err
	}

	items := root.FindAll("li", "class", "type-officer")

	results := make([]*UKSearchResult, 0, len(items))

	var wg sync.WaitGroup

	for _, item := range items {
		link := item.Find("a")
		if link.Error != nil {
			log.Warn().Err(link.Error).Str("html", item.HTML()).Msg("skip officer link")
			continue
		}
		appointmentsUrl := fmt.Sprintf("%s%s", websiteRootURL, link.Attrs()["href"])
		matches := rxOfficerID.FindStringSubmatch(appointmentsUrl)
		if len(matches) != 2 {
			log.Warn().Str("url", appointmentsUrl).Msg("failed to extract officer ID")
			continue
		}

		result := &UKSearchResult{
			URL:       appointmentsUrl,
			Companies: make([]UKCompany, 0, 1),
			ID:        matches[1],
		}

		results = append(results, result)
		wg.Add(1)
		go searchPeople(appointmentsUrl, result, &wg)
	}

	wg.Wait()
	log.Info().Msg("all search completed")

	return results, nil
}

func searchPeople(appointmentsUrl string, searchResult *UKSearchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := log.Logger.With().Str("url", appointmentsUrl).Logger()

	root, err := html.Parse(appointmentsUrl)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse page")
		return
	}

	if peopleName := root.Find("h1", "id", "officer-name"); peopleName.Error != nil {
		log.Warn().Err(peopleName.Error).Msg("failed to find officer name")
	} else {
		searchResult.Name = peopleName.Text()
	}

	if dateOfBirth := root.Find("dd", "id", "officer-date-of-birth-value"); dateOfBirth.Error != nil {
		logger.Warn().Err(dateOfBirth.Error).Msg("failed to find officer date of birth")
	} else if dateTime, err := time.Parse("January 2006", dateOfBirth.Text()); err == nil {
		searchResult.DateOfBirth = (*types.MonthDate)(&dateTime)
	}

	companyH2Blocs := root.FindAll("h2")
	logger.Debug().Msgf("%d company links found", len(companyH2Blocs))

	companyIndex := 1
	for _, h2Bloc := range companyH2Blocs {
		link := h2Bloc.Find("a")
		if link.Error != nil {
			logger.Warn().Err(link.Error).Str("html", h2Bloc.HTML()).Msg("skip company h2 bloc")
			continue
		}
		companyRelURL := link.Attrs()["href"]
		companyURL := fmt.Sprintf("%s%s", websiteRootURL, companyRelURL)
		companyID := companyRelURL[9:]
		companyName := strings.ReplaceAll(link.Text(), fmt.Sprintf(" (%s)", companyID), "")
		role := strings.Trim(
			h2Bloc.FindNextElementSibling().
				FindNextElementSibling().
				Find("dd", "id", fmt.Sprintf("appointment-type-value%d", companyIndex)).
				Text(),
			"\n ")
		roleAppointedOnRaw := strings.Trim(
			h2Bloc.FindNextElementSibling().
				FindNextElementSibling().
				Find("dd", "id", fmt.Sprintf("appointed-value%d", companyIndex)).
				Text(),
			"\n ")
		roleAppointedOn, _ := time.Parse("2 January 2006", roleAppointedOnRaw)
		status := strings.Trim(
			h2Bloc.FindNextElementSibling().
				Find("dd", "id", fmt.Sprintf("company-status-value-%d", companyIndex)).
				Text(),
			"\n ")
		address := strings.Trim(
			h2Bloc.FindNextElementSibling().
				Find("dd", "id", fmt.Sprintf("correspondence-address-value-%d", companyIndex)).
				Text(),
			"\n ")

		company := UKCompany{
			URL:             companyURL,
			ID:              companyID,
			Name:            companyName,
			Role:            role,
			RoleAppointedOn: types.Date(roleAppointedOn),
			Status:          status,
			Address:         address,
		}

		nationalityBloc := h2Bloc.
			FindNextElementSibling().
			FindNextElementSibling().
			FindNextElementSibling().
			Find("dd", "id", fmt.Sprintf("nationality-value%d", companyIndex))
		if nationalityBloc.Error == nil {
			nationality := strings.Trim(nationalityBloc.Text(), "\n ")
			company.Nationality = &nationality
		}
		countryBloc := h2Bloc.
			FindNextElementSibling().
			FindNextElementSibling().
			FindNextElementSibling().
			Find("dd", "id", fmt.Sprintf("country-of-residence-value%d", companyIndex))
		if countryBloc.Error == nil {
			country := strings.Trim(countryBloc.Text(), "\n ")
			company.CountryOfResidence = &country
		}
		occupationBloc := h2Bloc.
			FindNextElementSibling().
			FindNextElementSibling().
			FindNextElementSibling().
			Find("dd", "id", fmt.Sprintf("occupation-value-%d", companyIndex))
		if occupationBloc.Error == nil {
			occupation := strings.Trim(occupationBloc.Text(), "\n ")
			company.Occupation = &occupation
		}

		searchResult.Companies = append(searchResult.Companies, company)
		companyIndex++
	}
}
