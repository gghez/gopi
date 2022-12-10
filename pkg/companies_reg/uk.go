package companies_reg

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gghez/gopi/internal/html"
)

const websiteRootURL string = "https://find-and-update.company-information.service.gov.uk"

type UKCompany struct {
	ID                 string
	Name               string
	Role               string
	Status             string
	Address            string
	RoleAppointedOn    time.Time
	ResignedOn         *time.Time
	Nationality        *string
	CountryOfResidence *string
	Occupation         *string
	URL                string
}

type UKSearchResult struct {
	ID          string
	Name        string
	URL         string
	DateOfBirth time.Time
	Companies   []UKCompany
}

func SearchUK(query string) ([]*UKSearchResult, error) {
	log.Printf("search uk companies registry: %q", query)
	urlEncodedQuery := url.QueryEscape(query)

	url := fmt.Sprintf("%s/search?q=%s", websiteRootURL, urlEncodedQuery)
	root, err := html.ParseAndFind(url, "li", "class", "type-officer")
	if err != nil {
		return nil, err
	}

	links := root.FindAll("a")

	results := make([]*UKSearchResult, 0, len(links))

	var wg sync.WaitGroup

	for _, link := range links {
		appointmentsUrl := fmt.Sprintf("%s%s", websiteRootURL, link.Attrs()["href"])
		result := &UKSearchResult{URL: appointmentsUrl, Companies: make([]UKCompany, 0, 1)}
		results = append(results, result)
		wg.Add(1)
		go searchPeople(appointmentsUrl, result, &wg)
	}

	wg.Wait()
	log.Print("all search completed")

	return results, nil
}

func searchPeople(appointmentsUrl string, searchResult *UKSearchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	root, err := html.Parse(appointmentsUrl)
	if err != nil {
		log.Printf("failed to parse page: %q (error: %s)", appointmentsUrl, err)
		return
	}

	if peopleName := root.Find("h1", "id", "officer-name"); peopleName.Error != nil {
		log.Printf("failed to find officer name in page: %q (error: %s)", appointmentsUrl, peopleName.Error)
	} else {
		searchResult.Name = peopleName.Text()
	}

	companyBlocs := root.FindAll("h2")
	log.Printf("%d company links found", len(companyBlocs))

	for i, bloc := range companyBlocs {
		link := bloc.Find("a")
		if link.Error != nil {
			log.Printf("skipped bloc %d no link inside (error: %s)", i, link.Error)
			continue
		}
		companyRelURL := link.Attrs()["href"]
		companyURL := fmt.Sprintf("%s%s", websiteRootURL, companyRelURL)
		companyID := companyRelURL[9:]
		companyName := strings.ReplaceAll(link.Text(), fmt.Sprintf(" (%s)", companyID), "")
		companyRole := strings.Trim(bloc.FindNextElementSibling().FindNextElementSibling().Find("dd", "id", "appointment-type-value1").Text(), "\n ")
		companyRoleAppointedOnRaw := strings.Trim(bloc.FindNextElementSibling().FindNextElementSibling().Find("dd", "id", "appointed-value1").Text(), "\n ")
		companyRoleAppointedOn, _ := time.Parse("02 January 2006", companyRoleAppointedOnRaw)

		company := UKCompany{URL: companyURL, ID: companyID, Name: companyName, Role: companyRole, RoleAppointedOn: companyRoleAppointedOn}
		searchResult.Companies = append(searchResult.Companies, company)
	}
}
