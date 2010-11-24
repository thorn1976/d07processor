/*
	d07processor, Google Go Implementation
	2010-04-28, T. Frowein

	R 1.0.0
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	)

func cutFilterSymbol (currentFilter string, expression string) string {

	reg, err := regexp.Compile (expression);
	if  err != nil {

		fmt.Printf ("regex.Compile failed\n")
		os.Exit (1)
	}

	currentFilter = reg.ReplaceAllString (currentFilter, "")

	return currentFilter
}

func cutNegation (currentFilter string) string {

	return cutFilterSymbol (currentFilter, "^!")
}

func checkNegation (negation bool, result bool) bool {
	if negation == true {

		switch result {

			case true :

				result = false

			case false :

				result = true
		}
	}

	return result
}

func isFilterContent (feldnr int, inhalt string, filter map[string] string) (bool, bool) {

	var feldstring string = strconv.Itoa (feldnr + 1)
	var negation   bool   = false
	var done       bool   = false
	var result     bool   = false

	_, present := filter[feldstring]
	if present == true {

		currentFilter, ok := filter[feldstring]

		if ok != true {

			fmt.Printf ("could not access filter\n")
			os.Exit (1)
		}

		if currentFilter == "*" {

			result = true
			done   = true
		}

		if strings.Index (currentFilter, "§") > -1 && done == false {

			if strings.Index (currentFilter, "!") == 0 {

				negation = true
				currentFilter = cutNegation (currentFilter)
			}

			currentFilter = cutFilterSymbol (currentFilter, "^§")

			if currentFilter == inhalt {

				result = true
			}

			result = checkNegation (negation, result)
			done   = true
		}

		if strings.Index (currentFilter, "-") > -1 && done == false {

			if strings.Index (currentFilter, "!") == 0 {

				negation = true
				currentFilter = cutNegation (currentFilter)
			}

			parts   := strings.Split (currentFilter, "-", -1)
			von, _  := strconv.Atoi (parts[0])
			bis, _  := strconv.Atoi (parts[1])

			term, _ := strconv.Atoi (inhalt)

			if  von <= term && term <= bis {

				result = true
			}

			result = checkNegation (negation, result)
			done = true
		}

		if strings.Index (currentFilter, ",") > -1 && done == false {

			if strings.Index (currentFilter, "!") == 0 {

				negation = true
				currentFilter = cutNegation (currentFilter)
			}

			parts := strings.Split (currentFilter, ",", -1)

			for i := 0; i < len (parts); i++ {

				if parts[i] == inhalt {

					result = true
					break
				}
			}

			result = checkNegation (negation, result)
			done   = true
		}

		if strings.Index (currentFilter, "$") > -1 && done == false {

			if strings.Index (currentFilter, "!") == 0 {

				negation = true
				currentFilter = cutNegation (currentFilter)
			}

			currentFilter = cutFilterSymbol (currentFilter, "^\\$")

			reg, regErr := regexp.Compile (currentFilter)

			if regErr != nil {

				fmt.Printf ("failed to compile expression\n")
				os.Exit (1)
			}

			result = reg.MatchString (inhalt)

			result = checkNegation (negation, result)
			done   = true
		}
	}

	return present, result;
}

func main () {

	/*
		Kommandozeilen-Argumente auswerten
	*/

	if flag.NArg () < 2 {

		fmt.Printf ("Falsche Anzahl Argumente\n")
		fmt.Printf ("usage = d07proc bcp-Datei Filter-Datei\n")

		os.Exit (1)
	}

	var myFilter       = map[string] string {}
	var bcpFile string = flag.Arg (0)
	var cmdFile string = flag.Arg (1)

	/*
		Kommando-Datei öffnen und auslesen
	*/

	cfh, err := os.Open (cmdFile, os.O_RDONLY, 0666)

	if err != nil {
		fmt.Printf ("could not open file %s \n", cmdFile)

		os.Exit (1)
	}

	cmdReader := bufio.NewReader (cfh)

	zeile, readErr := cmdReader.ReadString (10)
	for readErr != os.EOF {

		match, regexErr := regexp.MatchString ("^#", zeile)
		if match == false && regexErr == nil {

			zeile  = strings.TrimSpace (zeile)
			dummy := strings.Split (zeile, "=", -1)

			myFilter[dummy[0]] = dummy[1]

		}

		zeile, readErr = cmdReader.ReadString (10)
	}

	err = cfh.Close ()

	/*
		Unload/BCP-Datei öffnen und auslesen
	*/

	bfh, err := os.Open (bcpFile, os.O_RDONLY, 0666)

	if err != nil {
		fmt.Printf ("could not open file %s \n", bcpFile)

		os.Exit (1)
	}

	bcpReader := bufio.NewReader (bfh)

	zeile, readErr = bcpReader.ReadString (10)
	for readErr != os.EOF {

		zeile  = strings.TrimSpace (zeile)
		dummy := strings.Split (zeile, "|", -1)

		var outZeile string = ""
		var isMatch  bool   = true
		for i := 0; i < len (dummy) && isMatch == true; i++ {

			var result, present bool

			present, result = isFilterContent (i, dummy[i], myFilter)

			if present == true {

				switch outZeile {

					case "" :
						outZeile = dummy[i]

					default :
						outZeile += "|" + dummy[i]
				}

				if result == false {

					isMatch = false
				}
			}
		}

		if isMatch == true {

			fmt.Printf ("%s\n", outZeile)
		}

		zeile, readErr = bcpReader.ReadString (10)
	}

	err = bfh.Close ()
}
