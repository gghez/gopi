package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gghez/gopi/pkg/companies_reg"
)

func main() {
	query := flag.String("q", "", "query terms")
	flag.Parse()

	results, err := companies_reg.SearchUK(*query)
	if err != nil {
		os.Exit(1)
	}

	for _, r := range results {
		fmt.Printf("name: %q\n", r.Name)
		fmt.Printf("id: %q\n", r.ID)
		fmt.Printf("companies:")
		if len(r.Companies) == 0 {
			fmt.Printf(" []\n")
		} else {
			for _, c := range r.Companies {
				fmt.Printf("\n- name: %q\n", c.Name)
				fmt.Printf("  id: %q\n", c.ID)
				fmt.Printf("  url: %q\n", c.URL)
				fmt.Printf("  role: %q\n", c.Role)
				fmt.Printf("  role_appointed_on: %q\n", c.RoleAppointedOn)
			}
		}

	}
}
