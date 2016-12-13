package mtparser

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var swiftChars = map[string]string{
	// n = Digits
	"n": "[0-9]",
	// d = Digits with decimal comma
	"d": "[0-9,]",
	// h = Uppercase hexadecimal
	"h": "[0-9A-F]",
	// a = Uppercase letters
	"a": "[A-Z]",
	// c = Uppercase alphanumeric
	"c": "[0-9A-Z]",
	// e = Space
	"e": "[ ]",
	// x = SWIFT character set
	"x": "[0-9A-Za-z/-?:().,'+\\r\\n ]",
	// y = Uppercase level A ISO 9735 characters
	"y": "[0-9A-Z.,\\-()/='+:?!\"%&*<>; ]",
	// z = SWIFT extended character set
	"z": "[0-9A-Za-z.,\\-()/+'=:?@_#\\r\\n {!\"%&*;<>]",
}

type reg struct {
	min string
	max string
	ptn string
	lns string
}

func (s *Parser) BodyValueStructured(k string) []string {
	var str string
	var rgx *regexp.Regexp

	if val, ok := FieldPatterns[k]; ok {
		str = regstrFromStructure(val["pattern"], val["fieldNames"])
		rgx = regexp.MustCompile(str)
	} else {
		return []string{}
	}

	// Range over 4 and parse the fields
	if val, ok := s.Map["4"][k]; ok {
		return rgx.FindStringSubmatch(val.Val)
	}

	return []string{}
}

func (s *Parser) ParseBody() error {
	var str string
	var rgx *regexp.Regexp

	if blk, ok := s.Map["4"]; ok {
		for k, v := range blk {
			if ptn, ok := FieldPatterns[k]; ok {
				if fld, ok := s.Map["4"][k]; ok {
					str = regstrFromStructure(ptn["pattern"], ptn["fieldNames"])
					rgx = regexp.MustCompile(str)
					mtc := rgx.FindStringSubmatch(v.Val)
					nms := rgx.SubexpNames()

					fld.Det = make(map[string]string)
					if len(nms) == len(mtc) {
						for i, name := range rgx.SubexpNames() {
							if i != 0 {
								fld.Det[name] = mtc[i]
							}
						}
					}

					s.Map["4"][k] = fld
				}
			}
		}
	}

	return nil
}

func TextRegexCompilation() {
	for k, v := range FieldPatterns {
		p := v["pattern"]
		rgx := regstrFromStructure(p, v["fieldNames"])
		fmt.Println("Field - ", k, " SWIFT - ", p, " REGEX - ", rgx)
		regexp.MustCompile(rgx)
	}
}

func regstrFromStructure(str string, keys string) string {
	var c rune
	var err error
	var r reg

	keys = strings.Replace(keys, "$", "", -1)
	keys = regexp.MustCompile("^[(]|[)]$|[ -]").ReplaceAllString(keys, "")
	kys := regexp.MustCompile("[)][(]").Split(keys, -1)

	mch := ""
	rgx := ""
	rdr := strings.NewReader(str)

	r.min = "0"

	for {
		c, _, err = rdr.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		switch c {
		case 'n', 'd', 'h', 'a', 'c', 'e', 'x', 'y', 'z':
			r.ptn = swiftChars[string(c)]
			mch += r.ptn + "{"
			if len(r.min) > 0 {
				mch += r.min + ","
			}
			mch += r.max + "}"

			if len(r.lns) > 0 {
				i, _ := strconv.Atoi(r.lns)
				mch += "(?:\\n" + mch + "){0," + strconv.Itoa(i-1) + "}"
			}

			if len(kys) > 0 {
				if len(kys[0]) > 0 {
					mch = "?P<" + kys[0] + ">" + mch
				}
				kys = kys[1:]
			}
			rgx += "(" + mch + ")"

			rgx = strings.Replace(rgx, ")?$", "\\r?\\n)?", -1)
			rgx = strings.Replace(rgx, "$", "\\r?\\n", -1)

			mch = ""
			r.max = ""
			r.min = "0"
			r.ptn = ""
			r.lns = ""
		case '!':
			r.min = ""
		case '[':
			rgx += "(?:"
		case ']':
			rgx += ")?"
		case '*':
			r.lns = r.max
			r.max = ""
			r.min = "0"
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			r.max += string(c)
		default:
			rgx += string(c)
		}
	}

	return rgx
}

var FieldPatterns = map[string]map[string]string{
	"12": {
		"pattern":    "3!n",
		"fieldNames": "",
	},
	"19": {
		"pattern":    "17d",
		"fieldNames": "(Amount)",
	},
	"20": {
		"pattern":    "16x",
		"fieldNames": "",
	},
	"21": {
		"pattern":    "16x",
		"fieldNames": "",
	},
	"23": {
		"pattern":    "16x",
		"fieldNames": "",
	},
	"25": {
		"pattern":    "35x",
		"fieldNames": "(Account)",
	},
	"28": {
		"pattern":    "5n[/2n]",
		"fieldNames": "(Page Number)(Indicator)",
	},
	"30": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"36": {
		"pattern":    "12d",
		"fieldNames": "(Rate)",
	},
	// I think this field has structure, should define and alternative for this
	"72": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"75": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"76": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"79": {
		// Currently this isn't supported by go regex. Needs to be 1k max chars
		// "pattern":    "35*50x",
		"pattern":    "20*50x",
		"fieldNames": "(Narrative)",
	},
	"11A": {
		"pattern":    ":4!c//3!a",
		"fieldNames": "(Qualifier)(Currency Code)",
	},
	"11R": {
		"pattern":    "3!n$6!n$[4!n6!n]",
		"fieldNames": "(MT Number)$(Date)$(Session Number)(ISN)",
	},
	"11S": {
		"pattern":    "3!n$6!n$[4!n6!n]",
		"fieldNames": "(MT Number)$(Date)$(Session Number)(ISN)",
	},
	"12A": {
		"pattern":    ":4!c/[8c]/30x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Instrument Code or Description)",
	},
	"12B": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Instrument Type Code)",
	},
	"12C": {
		"pattern":    ":4!c//6!c",
		"fieldNames": "(Qualifier)(CFI Code)",
	},
	"13A": {
		"pattern":    ":4!c//3!c",
		"fieldNames": "(Qualifier)(Number Identification)",
	},
	"13B": {
		"pattern":    ":4!c/[8c]/30x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Number)",
	},
	"13J": {
		"pattern":    ":4!c//5!c",
		"fieldNames": "(Qualifier)(Extended Number Id)",
	},
	"13K": {
		"pattern":    ":4!c//3!c/15d",
		"fieldNames": "(Qualifier)(Number Id)(Quantity)",
	},
	"16R": {
		"pattern":    "16c",
		"fieldNames": "",
	},
	"16S": {
		"pattern":    "16c",
		"fieldNames": "",
	},
	"17B": {
		"pattern":    ":4!c//1!a",
		"fieldNames": "(Qualifier)(Flag)",
	},
	"19A": {
		"pattern":    ":4!c//[N]3!a15d",
		"fieldNames": "(Qualifier)(Sign)(Currency Code)(Amount)",
	},
	"19B": {
		"pattern":    ":4!c//3!a15d",
		"fieldNames": "(Qualifier)(Currency Code)(Amount)",
	},
	"20C": {
		"pattern":    ":4!c//16x",
		"fieldNames": "(Qualifier)(Reference)",
	},
	"20D": {
		"pattern":    ":4!c//25x",
		"fieldNames": "(Qualifier)(Reference)",
	},
	"22F": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Indicator)",
	},
	"22H": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Indicator)",
	},
	"23B": {
		"pattern":    "4!c",
		"fieldNames": "(Function)",
	},
	"23E": {
		"pattern":    "4!c[/30x]",
		"fieldNames": "(Function)(Additional Information)",
	},
	"23G": {
		"pattern":    "4!c[/4!c]",
		"fieldNames": "(Function)(Subfunction)",
	},
	"24B": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Reason Code)",
	},
	"25D": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Status Code)",
	},
	"26H": {
		"pattern":    "16x",
		"fieldNames": "",
	},
	"28D": {
		"pattern":    "5n/5n",
		"fieldNames": "(Message Index)(Total)",
	},
	"28E": {
		"pattern":    "5n/4!c",
		"fieldNames": "(Page Number)(Continuation Indicator)",
	},
	"29A": {
		"pattern":    "4*35x",
		"fieldNames": "(Narrative)",
	},
	"29B": {
		"pattern":    "4*35x",
		"fieldNames": "(Narrative)",
	},
	"31C": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"31E": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"31F": {
		"pattern":    "6!n[/6!n][//35x]",
		"fieldNames": "(Date)(Period Date)(Period Details)",
	},
	"31L": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"31P": {
		"pattern":    "6!n[29x]",
		"fieldNames": "(Date)(Place)",
	},
	"31S": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"31X": {
		"pattern":    "[6!n[4!n]][7!a]",
		"fieldNames": "(Date)(Time)(Code)",
	},
	"32A": {
		"pattern":    "6!n3!a15d",
		"fieldNames": "(Date)(Currency)(Amount)",
	},
	"32B": {
		"pattern":    "3!a15d",
		"fieldNames": "(Currency)(Amount)",
	},
	"32C": {
		"pattern":    "6!n3!a15d",
		"fieldNames": "(Date)(Currency)(Amount)",
	},
	"32D": {
		"pattern":    "6!n3!a15d",
		"fieldNames": "(Date)(Currency)(Amount)",
	},
	"32G": {
		"pattern":    "3!a15d",
		"fieldNames": "(Currency)(Amount)",
	},
	"32M": {
		"pattern":    "3!a15d",
		"fieldNames": "(Currency)(Amount)",
	},
	"33B": {
		"pattern":    "3!a15d",
		"fieldNames": "(Code)(Amount)",
	},
	"33S": {
		"pattern":    "3!a15d",
		"fieldNames": "(Code)(Amount)",
	},
	"33T": {
		"pattern":    "3!a15d",
		"fieldNames": "(Currency)(Price)",
	},
	"34A": {
		"pattern":    "6!n3!a15d",
		"fieldNames": "(Date)(Currency)(Amount)",
	},
	"34B": {
		"pattern":    "3!a15d",
		"fieldNames": "(Currency)(Amount)",
	},
	"35A": {
		"pattern":    "3!a15d",
		"fieldNames": "(Type)(Quantity)",
	},
	"35B": {
		"pattern":    "[ISIN1!e12!c]$[4*35x]",
		"fieldNames": "(Identification of Security)$(Description of Security)",
	},
	"35C": {
		"pattern":    "3!c",
		"fieldNames": "(Number)",
	},
	"35D": {
		"pattern":    "6!n",
		"fieldNames": "(Date)",
	},
	"35E": {
		"pattern":    "6*50x",
		"fieldNames": "(Narrative)",
	},
	"35F": {
		// Not supported by golangs regex lib
		// "pattern":    "35*50x",
		"pattern":    "20*50x",
		"fieldNames": "(Narrative)",
	},
	"35H": {
		"pattern":    "[N]3!a15d",
		"fieldNames": "(Sign)(Currency)(Quantity)",
	},
	"35L": {
		"pattern":    "4*35x",
		"fieldNames": "(Narrative)",
	},
	"35N": {
		"pattern":    "3!a15d",
		"fieldNames": "(Type)(Quantity)",
	},
	"35S": {
		"pattern":    "3!a15d",
		"fieldNames": "(Type)(Quantity)",
	},
	"35U": {
		"pattern":    "3!a15d[1!a]",
		"fieldNames": "(Currency)(Price)(Period)",
	},
	"36B": {
		"pattern":    ":4!c//4!c/15d",
		"fieldNames": "(Qualifier)(Quantity Type Code)(Quantity)",
	},
	"36C": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Quantity Code)",
	},
	"36E": {
		"pattern":    ":4!c//4!c/[N]15d",
		"fieldNames": "(Qualifier)(Quantity Type Code)(Sign)(Quantity)",
	},
	"37A": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37B": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37C": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37D": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37E": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37F": {
		"pattern":    "12d[//6!n1!a3n][/16x]",
		"fieldNames": "(Rate)(End Date)(Period)(Number)(Information)",
	},
	"37J": {
		"pattern":    "12d",
		"fieldNames": "(Rate)",
	},
	"50H": {
		"pattern":    "/34x$4*35x",
		"fieldNames": "(Account)$(Name and Address)",
	},
	"50K": {
		"pattern":    "[/34x]$4*35x",
		"fieldNames": "(Account)$(Name and Address)",
	},
	"52A": {
		"pattern":    "[/1!a][/34x]$4!a2!a2!c[3!c]",
		"fieldNames": "(Party Identifier)$(Identifier Code)",
	},
	"52D": {
		"pattern":    "[/1!a][/34x]$4*35x",
		"fieldNames": "(Party Identifier)$(Name and Address)",
	},
	"53A": {
		"pattern":    "[/1!a][/34x]$4!a2!a2!c[3!c]",
		"fieldNames": "(Party Identifier)$(Identifier Code)",
	},
	"53C": {
		"pattern":    "/34x",
		"fieldNames": "(Account)",
	},
	"53D": {
		"pattern":    "[/1!a][/34x]$4*35x",
		"fieldNames": "(Party Identifier)$(Name and Address)",
	},
	"57A": {
		"pattern":    "[/1!a][/34x]$4!a2!a2!c[3!c]",
		"fieldNames": "(Party Identifier)$(Identifier Code)",
	},
	"57B": {
		"pattern":    "[/1!a][/34x]$[35x]",
		"fieldNames": "(Party Identifier)$(Location)",
	},
	"57D": {
		"pattern":    "[/1!a][/34x]$4*35x",
		"fieldNames": "(Party Identifier)$(Name and Address)",
	},
	"59": {
		"pattern":    "[/34x]$4*35x",
		"fieldNames": "(Account)$(Name and Address)",
	},
	"67A": {
		"pattern":    "6!n[/6!n]",
		"fieldNames": "(Date 1)(Date 2)",
	},
	"69A": {
		"pattern":    ":4!c//8!n/8!n",
		"fieldNames": "(Qualifier)(Date)(Date)",
	},
	"69B": {
		"pattern":    ":4!c//8!n6!n/8!n6!n",
		"fieldNames": "(Qualifier)(Date)(Time)(Date)(Time)",
	},
	"69C": {
		"pattern":    ":4!c//8!n/4!c",
		"fieldNames": "(Qualifier)(Date)(Date Code)",
	},
	"69D": {
		"pattern":    ":4!c//8!n6!n/4!c",
		"fieldNames": "(Qualifier)(Date)(Time)(Date Code)",
	},
	"69E": {
		"pattern":    ":4!c//4!c/8!n",
		"fieldNames": "(Qualifier)(Date Code)(Date)",
	},
	"69F": {
		"pattern":    ":4!c//4!c/8!n6!n",
		"fieldNames": "(Qualifier)(Date Code)(Date)(Time)",
	},
	"69J": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Date Code)",
	},
	"70": {
		"pattern":    "4*35x",
		"fieldNames": "(Narrative)",
	},
	"70C": {
		"pattern":    ":4!c//4*35x",
		"fieldNames": "(Qualifier)(Narrative)",
	},
	"70D": {
		"pattern":    ":4!c//6*35x",
		"fieldNames": "(Qualifier)(Narrative)",
	},
	"70E": {
		"pattern":    ":4!c//10*35x",
		"fieldNames": "(Qualifier)(Narrative)",
	},
	"70F": {
		// Currently this is not supported by golang regex. Need to work out a better implementation
		// "pattern":    ":4!c//8000z",
		"pattern":    ":4!c//800z",
		"fieldNames": "(Qualifier)(Narrative)",
	},
	"70G": {
		"pattern":    ":4!c//10*35z",
		"fieldNames": "(Qualifier)(Narrative)",
	},
	"71A": {
		"pattern":    "3!a",
		"fieldNames": "(Details Of Charges)",
	},
	"71B": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"71C": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"71F": {
		"pattern":    "3!a15d",
		"fieldNames": "(Code)(Amount)",
	},
	"77A": {
		"pattern":    "20*35x",
		"fieldNames": "(Narrative)",
	},
	"77D": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"77E": {
		"pattern":    "73x$[n*78x]",
		"fieldNames": "(Text)$(Text)",
	},
	"80C": {
		"pattern":    "6*35x",
		"fieldNames": "(Narrative)",
	},
	"83A": {
		"pattern":    "[/1!a][/34x]$4!a2!a2!c[3!c]",
		"fieldNames": "(Party Identifier)$(Identifier Code)",
	},
	"83C": {
		"pattern":    "/34x",
		"fieldNames": "(Account)",
	},
	"83D": {
		"pattern":    "[/1!a][/34x]$4*35x",
		"fieldNames": "(Party Identifier)$(Name and Address)",
	},
	"87A": {
		"pattern":    "[/1!a][/34x]$4!a2!a2!c[3!c]",
		"fieldNames": "(Party Identifier)$(Identifier Code)",
	},
	"87D": {
		"pattern":    "[/1!a][/34x]$4*35x",
		"fieldNames": "(Party Identifier)$(Name and Address)",
	},
	"90A": {
		"pattern":    ":4!c//4!c/15d",
		"fieldNames": "(Qualifier)(Percentage Type Code)(Price)",
	},
	"90B": {
		"pattern":    ":4!c//4!c/3!a15d",
		"fieldNames": "(Qualifier)(Amount Type Code)(Currency Code)(Price)",
	},
	"90E": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Price Code)",
	},
	"90F": {
		"pattern":    ":4!c//4!c/3!a15d/4!c/15d",
		"fieldNames": "(Qualifier)(Amount Type Code)(Currency Code)(Amount)(Quantity Type Code)(Quantity)",
	},
	"90J": {
		"pattern":    ":4!c//4!c/3!a15d/3!a15d",
		"fieldNames": "(Qualifier)(Amount Type Code)(Currency Code)(Amount)(Currency Code)(Amount)",
	},
	"90K": {
		"pattern":    ":4!c//15d",
		"fieldNames": "(Qualifier)(Index Points)",
	},
	"90L": {
		"pattern":    ":4!c//[N]15d",
		"fieldNames": "(Qualifier)(Sign)(Index Points)",
	},
	"92A": {
		"pattern":    ":4!c//[N]15d",
		"fieldNames": "(Qualifier)(Sign)(Rate)",
	},
	"92B": {
		"pattern":    ":4!c//3!a/3!a/15d",
		"fieldNames": "(Qualifier)(First Currency Code)(Second Currency Code)(Rate)",
	},
	"92C": {
		"pattern":    ":4!c/[8c]/24x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Rate Name)",
	},
	"92D": {
		"pattern":    ":4!c//15d/15d",
		"fieldNames": "(Qualifier)(Quantity)(Quantity)",
	},
	"92F": {
		"pattern":    ":4!c//3!a15d",
		"fieldNames": "(Qualifier)(Currency Code)(Amount)",
	},
	"92J": {
		"pattern":    ":4!c/[8c]/4!c/3!a15d[/4!c]",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Rate Type Code)(Currency Code)(Amount)(Rate Status)",
	},
	"92K": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Rate Type Code)",
	},
	"92L": {
		"pattern":    ":4!c//3!a15d/3!a15d",
		"fieldNames": "(Qualifier)(First Currency Code)(Amount)(Second Currency Code)(Amount)",
	},
	"92M": {
		"pattern":    ":4!c//3!a15d/15d",
		"fieldNames": "(Qualifier)(Currency Code)(Amount)(Quantity)",
	},
	"92N": {
		"pattern":    ":4!c//15d/3!a15d",
		"fieldNames": "(Qualifier)(Quantity)(Currency Code)(Amount)",
	},
	"92P": {
		"pattern":    ":4!c//15d",
		"fieldNames": "(Qualifier)(Index Points)",
	},
	"92R": {
		"pattern":    ":4!c/[8c]/4!c/15d",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Rate Type Code)(Rate)",
	},
	"93A": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Sub-balance Type)",
	},
	"93B": {
		"pattern":    ":4!c/[8c]/4!c/[N]15d",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Quantity Type Code)(Sign)(Balance)",
	},
	"93C": {
		"pattern":    ":4!c//4!c/4!c/[N]15d",
		"fieldNames": "(Qualifier)(Quantity Type Code)(Balance Type Code)(Sign)(Balance)",
	},
	"93D": {
		"pattern":    ":4!c//[N]15d",
		"fieldNames": "(Qualifier)(Sign)(Balance)",
	},
	"94B": {
		"pattern":    ":4!c/[8c]/4!c[/30x]",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Place Code)(Narrative)",
	},
	"94C": {
		"pattern":    ":4!c//2!a",
		"fieldNames": "(Qualifier)(Country Code)",
	},
	"94D": {
		"pattern":    ":4!c//[2!a]/35x",
		"fieldNames": "(Qualifier)(Country Code)(Place)",
	},
	"94E": {
		"pattern":    ":4!c//10*35x",
		"fieldNames": "(Qualifier)(Address)",
	},
	"94F": {
		"pattern":    ":4!c//4!c/4!a2!a2!c[3!c]",
		"fieldNames": "(Qualifier)(Place Code)(Identifier Code)",
	},
	"94G": {
		"pattern":    ":4!c//2*35x",
		"fieldNames": "(Qualifier)(Address)",
	},
	"94H": {
		"pattern":    ":4!c//4!a2!a2!c[3!c]",
		"fieldNames": "(Qualifier)(Identifier Code)",
	},
	"95C": {
		"pattern":    ":4!c//2!a",
		"fieldNames": "(Qualifier)(Country Code)",
	},
	"95P": {
		"pattern":    ":4!c//4!a2!a2!c[3!c]",
		"fieldNames": "(Qualifier)(Identifier Code)",
	},
	"95Q": {
		"pattern":    ":4!c//4*35x",
		"fieldNames": "(Qualifier)(Name and Address)",
	},
	"95R": {
		"pattern":    ":4!c/8c/34x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Proprietary Code)",
	},
	"95S": {
		"pattern":    ":4!c/[8c]/4!c/2!a/30x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Type of ID)(Country Code)(Alternate ID)",
	},
	"95T": {
		"pattern":    ":4!c//2*35x",
		"fieldNames": "(Qualifier)(Name)",
	},
	"95U": {
		"pattern":    ":4!c//3*35x",
		"fieldNames": "(Qualifier)(Name)",
	},
	"95V": {
		"pattern":    ":4!c//10*35x",
		"fieldNames": "(Qualifier)(Name and Address)",
	},
	"97A": {
		"pattern":    ":4!c//35x",
		"fieldNames": "(Qualifier)(Account Number)",
	},
	"97B": {
		"pattern":    ":4!c/[8c]/4!c/35x",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Account Type Code)(Account Number)",
	},
	"97C": {
		"pattern":    ":4!c//4!c",
		"fieldNames": "(Qualifier)(Account Code)",
	},
	"97E": {
		"pattern":    ":4!c//34x",
		"fieldNames": "(Qualifier)(International Bank Account Number)",
	},
	"98A": {
		"pattern":    ":4!c//8!n",
		"fieldNames": "(Qualifier)(Date)",
	},
	"98B": {
		"pattern":    ":4!c/[8c]/4!c",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Date Code)",
	},
	"98C": {
		"pattern":    ":4!c//8!n6!n",
		"fieldNames": "(Qualifier)(Date)(Time)",
	},
	"98E": {
		"pattern":    ":4!c//8!n6!n[,3n][/[N]2!n[2!n]]",
		"fieldNames": "(Qualifier)(Date)(Time)(Decimals)(UTC Sign)(UTC Indicator)",
	},
	"98F": {
		"pattern":    ":4!c/[8c]/4!c6!n",
		"fieldNames": "(Qualifier)(Data Source Scheme)(Date Code)(Time)",
	},
	"99A": {
		"pattern":    ":4!c//[N]3!n",
		"fieldNames": "(Qualifier)(Sign)(Number)",
	},
	"99B": {
		"pattern":    ":4!c//3!n",
		"fieldNames": "(Qualifier)(Number)",
	},
	"99C": {
		"pattern":    ":4!c//6!n",
		"fieldNames": "(Qualifier)(Number)",
	},
}
