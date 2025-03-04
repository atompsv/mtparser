#  Quick Start Guide
##  Create a text file with an MT103 message.
Save the MT103 message in a text file, e.g., input.txt:
```
{1:F01AAAAGRA0AXXX0057000289}{2:O1030919010321BBBBGRA0AXXX00570001710103210920N}{4:
:20:5387354
:23B:CRED
:23E:PHOB/20.527.19.60
:32A:000526USD1101,50
:33B:USD1121,50
:50K:FRANZ HOLZAPFEL GMBH
VIENNA
:52A:BKAUATWW
:59:723491524
C. KLEIN
BLOEMENGRACHT 15
AMSTERDAM
:71A:SHA
:71F:USD10,
:71F:USD10,
:72:/INS/CHASUS33
-}{5:{MAC:75D138E4}{CHK:DE1B0D71FA96}}
```

## Run the following Go code to parse the file
```go

	// Open the MT103 file
	file, err := os.Open("input.txt")
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer file.Close()

	// Create a buffered reader
	r := bufio.NewReader(file)

	// Initialize the MT parser
	psr := mtparser.New(r)

	// Parse the MT103 message
	if err = psr.Parse(); err != nil {
		log.Fatalln("Error parsing MT103 message:", err)
	}

	// Print the parsed data
	fmt.Println(psr.Map)
```
   