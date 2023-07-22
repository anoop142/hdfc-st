# hdfc-st
A command-line tool to print data from HDFC bank CSV statement files.

## Usage:
```
./hdfc-st -f file | - [-d text to match] [-x text to exclude] [-on dd/mm/yyyy] | [-from -after dd/mm/yyyy]
  -cred
    	print credits
  -d value
    	description to match
  -deb
    	print debits
  -f string
    	statement text file | - stdin
  -from value
    	transactions from specified date
  -net
    	print net
  -on value
    	transactions on specified date
  -to value
    	transactions till specified date
  -x value
    	description to exclude

```

## Features

* Filter by
    * Matching description
    * Excluding description
    * On specific date
    * Between dates

* Table formatted
* Print only credits and debits


## Example

```
hdfc-st -f statement.txt -from 20/07/2023 -to 22/07/2023 

```


## Building
```
go build
```
