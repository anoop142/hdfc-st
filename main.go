/*
hdfc-st:- A tool to display info from HDFC statement CSV file

Anoop S

2023

*/

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

const (
	DESCRIPTION_MAX_LEN = 15
)

type config struct {
	file         string
	descriptions []string
	deb          bool
	cred         bool
	net          bool
	fromDate     time.Time
	toDate       time.Time
	onDate       time.Time
	exclude      []string
}

type statement struct {
	Date        time.Time
	DateFmt     string
	Description string
	Debit       float64
	Credit      float64
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Printf("%s -f file | - [-d text to match] [-x text to exclude] [-on dd/mm/yyyy] | [-from -after dd/mm/yyyy]\n", os.Args[0])
	flag.PrintDefaults()

}

func (cfg config) getStatements() ([]statement, error) {
	return nil, nil
}

func fixDate(d string) string {
	tokens := strings.Split(d, "/")
	return fmt.Sprintf("%s/%s/20%s", tokens[0], tokens[1], tokens[2])
}

func getStatement(line string) (statement, error) {
	var st statement

	tokens := strings.Split(line, ",")

	if len(tokens) < 6 {
		return st, errors.New("error parsing statement file")
	}

	dateFmt := fixDate(strings.TrimSpace(tokens[0]))

	date, err := time.Parse("02/01/2006", fixDate(strings.TrimSpace(tokens[0])))
	if err != nil {
		return st, err
	}

	description := strings.TrimSpace(tokens[1])

	debit, err := strconv.ParseFloat(strings.TrimSpace(tokens[3]), 64)
	if err != nil {
		return st, err
	}

	credit, err := strconv.ParseFloat(strings.TrimSpace(tokens[4]), 64)
	if err != nil {
		return st, err
	}

	st = statement{
		DateFmt:     dateFmt,
		Date:        date,
		Description: description,
		Debit:       debit,
		Credit:      credit,
	}

	return st, nil

}

func matchDescription(descriptions []string, description string) bool {
	for _, v := range descriptions {
		if strings.Contains(description, v) {
			return true
		}
	}

	return false
}

func (cfg config) filterStatement(st *statement) bool {
	// filter by decription
	if len(cfg.descriptions) != 0 && !matchDescription(cfg.descriptions, st.Description) {
		return false
	}

	if len(cfg.exclude) != 0 && matchDescription(cfg.exclude, st.Description) {
		return false
	}

	if !cfg.onDate.IsZero() && !st.Date.Equal(cfg.onDate) {
		return false
	}

	if !cfg.fromDate.IsZero() {
		if !st.Date.Equal(cfg.fromDate) && !st.Date.After(cfg.fromDate) {
			return false
		}
	}

	if !cfg.toDate.IsZero() {
		if !st.Date.Equal(cfg.toDate) && !st.Date.Before(cfg.toDate) {
			return false
		}
	}

	return true

}

func float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func calculateNetFmt(debt, credit float64) string{
	net := debt - credit
	if net < 0{
		// add '+'
		return "+" + float64ToString(math.Abs(net))
	}

	return "-" + float64ToString(net)
}

func main() {
	var cfg config
	var credit float64
	var debit float64
	var count int

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Date", "Description", "Debit", "Credit"})

	flag.StringVar(&cfg.file, "f", "", "statement text file | - stdin")
	flag.BoolVar(&cfg.cred, "cred", false, "print credits")
	flag.BoolVar(&cfg.deb, "deb", false, "print debits")
	flag.BoolVar(&cfg.net, "net", false, "print net")
	flag.Func("d", "descriptions to match", func(value string) error {
		for _, v := range strings.Split(value, ",") {
			cfg.descriptions = append(cfg.descriptions, strings.ToUpper(strings.TrimSpace(v)))
		}
		return nil
	})

	flag.Func("x", "descriptions to exclude", func(value string) error {
		for _, v := range strings.Split(value, ",") {
			cfg.exclude = append(cfg.exclude, strings.ToUpper(strings.TrimSpace(v)))
		}

		return nil
	})

	flag.Func("on", "transactions on specified date", func(value string) error {
		date, err := time.Parse("02/01/2006", value)
		cfg.onDate = date
		return err
	})

	flag.Func("from", "transactions from specified date", func(value string) error {
		date, err := time.Parse("02/01/2006", value)
		cfg.fromDate = date
		if !cfg.onDate.IsZero() {
			return errors.New("cannot use -from along with -on")
		}
		return err
	})

	flag.Func("to", "transactions till specified date", func(value string) error {
		date, err := time.Parse("02/01/2006", value)
		cfg.toDate = date
		if !cfg.onDate.IsZero() {
			return errors.New("cannot use -to along with -on")
		}
		return err
	})

	flag.Usage = printUsage
	flag.Parse()

	var fileReader *os.File

	switch cfg.file{
	case "":
		printUsage()
		os.Exit(1)
	case "-":
		fileReader = os.Stdin
	
	default:
		f, err := os.Open(cfg.file)
		
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		
		defer f.Close()
		fileReader = f
	}

		scanner := bufio.NewScanner(fileReader)

	// find header and skip it
	foundHeader := false
	foundStart := false

	for scanner.Scan() {
		if !foundHeader && !foundStart{
			if strings.Contains(scanner.Text(), "Date"){
				foundHeader = true
				continue

			}else if scanner.Text() == ""{
				continue
			}
		}

		line := scanner.Text()

		st, err := getStatement(line)
		if err != nil {
			// keep going even on err
			fmt.Fprintln(os.Stderr, err)
			continue
		}



		foundStart = true

		// didn't pass filter, skip
		if !cfg.filterStatement(&st) {
			continue
		}

		count++

		// calculate debit and credit
		credit += st.Credit
		debit += st.Debit

		// skip, just calculate debit and credit
		if cfg.cred || cfg.deb || cfg.net {
			continue
		}

		// add to table
		table.Append([]string{
			st.DateFmt,
			st.Description[:DESCRIPTION_MAX_LEN],
			float64ToString(st.Debit),
			float64ToString(st.Credit),
		})

	}

	if count == 0 {
		fmt.Fprint(os.Stderr, "No match found!\n")
		os.Exit(1)
	}

	if cfg.deb {
		fmt.Println(float64ToString(debit))
	} else if cfg.cred {
		fmt.Println(float64ToString(credit))
	} else if cfg.net {
		fmt.Println(float64ToString((math.Abs(debit - credit))))
	} else {
		table.SetFooter([]string{"NET", calculateNetFmt(debit, credit), float64ToString(debit), float64ToString(credit)})
		table.Render()
	}

}
