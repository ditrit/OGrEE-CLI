// Code generated by goyacc - DO NOT EDIT.

package main

import __yyfmt__ "fmt"

import (
	cmd "cli/controllers"
	"strconv"
	"strings"
)

func resMap(x *string) map[string]interface{} {
	resarr := strings.Split(*x, "=")
	res := make(map[string]interface{})
	attrs := make(map[string]string)

	for i := 0; i+1 < len(resarr); {
		if i+1 < len(resarr) {
			switch resarr[i] {
			case "id", "name", "category", "parentID",
				"description", "domain", "parentid", "parentId":
				res[resarr[i]] = resarr[i+1]

			default:
				attrs[resarr[i]] = resarr[i+1]
			}
			i += 2
		}
	}
	res["attributes"] = attrs
	return res
}

func replaceOCLICurrPath(x string) string {
	return strings.Replace(x, "_", cmd.State.CurrPath, 1)
}

type yySymType struct {
	yys  int
	n    int
	s    string
	sarr []string
}

type yyXError struct {
	state, xsym int
}

const (
	yyDefault        = 57404
	yyEofCode        = 57344
	TOKEN_ATTR       = 57356
	TOKEN_ATTRSPEC   = 57389
	TOKEN_BASHTYPE   = 57364
	TOKEN_BLDG       = 57350
	TOKEN_CD         = 57370
	TOKEN_CLR        = 57372
	TOKEN_CMDFLAG    = 57366
	TOKEN_CMDS       = 57400
	TOKEN_COMMA      = 57398
	TOKEN_CREATE     = 57359
	TOKEN_DELETE     = 57362
	TOKEN_DEREF      = 57403
	TOKEN_DEVICE     = 57353
	TOKEN_DOC        = 57369
	TOKEN_DOT        = 57399
	TOKEN_EQUAL      = 57365
	TOKEN_EXIT       = 57368
	TOKEN_GET        = 57360
	TOKEN_GREP       = 57373
	TOKEN_LBRAC      = 57396
	TOKEN_LS         = 57374
	TOKEN_LSBLDG     = 57379
	TOKEN_LSDEV      = 57382
	TOKEN_LSOG       = 57376
	TOKEN_LSRACK     = 57381
	TOKEN_LSROOM     = 57380
	TOKEN_LSSITE     = 57378
	TOKEN_LSSUBDEV   = 57383
	TOKEN_LSSUBDEV1  = 57384
	TOKEN_LSTEN      = 57377
	TOKEN_NUM        = 57346
	TOKEN_OCBLDG     = 57385
	TOKEN_OCDEL      = 57358
	TOKEN_OCDEV      = 57386
	TOKEN_OCPSPEC    = 57394
	TOKEN_OCRACK     = 57387
	TOKEN_OCROOM     = 57388
	TOKEN_OCSDEV     = 57392
	TOKEN_OCSDEV1    = 57393
	TOKEN_OCSITE     = 57390
	TOKEN_OCTENANT   = 57391
	TOKEN_PLUS       = 57357
	TOKEN_PWD        = 57371
	TOKEN_RACK       = 57352
	TOKEN_RBRAC      = 57397
	TOKEN_ROOM       = 57351
	TOKEN_SEARCH     = 57363
	TOKEN_SELECT     = 57395
	TOKEN_SITE       = 57349
	TOKEN_SLASH      = 57367
	TOKEN_SUBDEVICE  = 57354
	TOKEN_SUBDEVICE1 = 57355
	TOKEN_TEMPLATE   = 57401
	TOKEN_TENANT     = 57348
	TOKEN_TREE       = 57375
	TOKEN_UPDATE     = 57361
	TOKEN_VAR        = 57402
	TOKEN_WORD       = 57347
	yyErrCode        = 57345

	yyMaxDepth = 200
	yyTabOfs   = -98
)

var (
	yyPrec = map[int]int{}

	yyXLAT = map[int]int{
		57344: 0,  // $end (103x)
		57347: 1,  // TOKEN_WORD (77x)
		57399: 2,  // TOKEN_DOT (53x)
		57389: 3,  // TOKEN_ATTRSPEC (44x)
		57367: 4,  // TOKEN_SLASH (42x)
		57424: 5,  // P1 (35x)
		57346: 6,  // TOKEN_NUM (34x)
		57423: 7,  // P (32x)
		57356: 8,  // TOKEN_ATTR (28x)
		57358: 9,  // TOKEN_OCDEL (25x)
		57357: 10, // TOKEN_PLUS (25x)
		57422: 11, // ORIENTN (24x)
		57426: 12, // WORDORNUM (23x)
		57394: 13, // TOKEN_OCPSPEC (15x)
		57407: 14, // F (5x)
		57365: 15, // TOKEN_EQUAL (5x)
		57397: 16, // TOKEN_RBRAC (4x)
		57350: 17, // TOKEN_BLDG (3x)
		57353: 18, // TOKEN_DEVICE (3x)
		57352: 19, // TOKEN_RACK (3x)
		57351: 20, // TOKEN_ROOM (3x)
		57349: 21, // TOKEN_SITE (3x)
		57348: 22, // TOKEN_TENANT (3x)
		57406: 23, // E (2x)
		57408: 24, // GETOBJS (2x)
		57370: 25, // TOKEN_CD (2x)
		57359: 26, // TOKEN_CREATE (2x)
		57362: 27, // TOKEN_DELETE (2x)
		57360: 28, // TOKEN_GET (2x)
		57396: 29, // TOKEN_LBRAC (2x)
		57374: 30, // TOKEN_LS (2x)
		57376: 31, // TOKEN_LSOG (2x)
		57354: 32, // TOKEN_SUBDEVICE (2x)
		57355: 33, // TOKEN_SUBDEVICE1 (2x)
		57375: 34, // TOKEN_TREE (2x)
		57361: 35, // TOKEN_UPDATE (2x)
		57405: 36, // BASH (1x)
		57409: 37, // K (1x)
		57410: 38, // NT_CREATE (1x)
		57411: 39, // NT_DEL (1x)
		57412: 40, // NT_GET (1x)
		57413: 41, // NT_UPDATE (1x)
		57414: 42, // OCCHOOSE (1x)
		57415: 43, // OCCR (1x)
		57416: 44, // OCDEL (1x)
		57417: 45, // OCDOT (1x)
		57418: 46, // OCGET (1x)
		57419: 47, // OCLISYNTX (1x)
		57420: 48, // OCSEL (1x)
		57421: 49, // OCUPDATE (1x)
		57425: 50, // Q (1x)
		57427: 51, // start (1x)
		57372: 52, // TOKEN_CLR (1x)
		57400: 53, // TOKEN_CMDS (1x)
		57398: 54, // TOKEN_COMMA (1x)
		57403: 55, // TOKEN_DEREF (1x)
		57369: 56, // TOKEN_DOC (1x)
		57368: 57, // TOKEN_EXIT (1x)
		57373: 58, // TOKEN_GREP (1x)
		57379: 59, // TOKEN_LSBLDG (1x)
		57382: 60, // TOKEN_LSDEV (1x)
		57381: 61, // TOKEN_LSRACK (1x)
		57380: 62, // TOKEN_LSROOM (1x)
		57378: 63, // TOKEN_LSSITE (1x)
		57383: 64, // TOKEN_LSSUBDEV (1x)
		57384: 65, // TOKEN_LSSUBDEV1 (1x)
		57377: 66, // TOKEN_LSTEN (1x)
		57385: 67, // TOKEN_OCBLDG (1x)
		57386: 68, // TOKEN_OCDEV (1x)
		57387: 69, // TOKEN_OCRACK (1x)
		57388: 70, // TOKEN_OCROOM (1x)
		57390: 71, // TOKEN_OCSITE (1x)
		57391: 72, // TOKEN_OCTENANT (1x)
		57371: 73, // TOKEN_PWD (1x)
		57395: 74, // TOKEN_SELECT (1x)
		57401: 75, // TOKEN_TEMPLATE (1x)
		57402: 76, // TOKEN_VAR (1x)
		57404: 77, // $default (0x)
		57345: 78, // error (0x)
		57364: 79, // TOKEN_BASHTYPE (0x)
		57366: 80, // TOKEN_CMDFLAG (0x)
		57392: 81, // TOKEN_OCSDEV (0x)
		57393: 82, // TOKEN_OCSDEV1 (0x)
		57363: 83, // TOKEN_SEARCH (0x)
	}

	yySymNames = []string{
		"$end",
		"TOKEN_WORD",
		"TOKEN_DOT",
		"TOKEN_ATTRSPEC",
		"TOKEN_SLASH",
		"P1",
		"TOKEN_NUM",
		"P",
		"TOKEN_ATTR",
		"TOKEN_OCDEL",
		"TOKEN_PLUS",
		"ORIENTN",
		"WORDORNUM",
		"TOKEN_OCPSPEC",
		"F",
		"TOKEN_EQUAL",
		"TOKEN_RBRAC",
		"TOKEN_BLDG",
		"TOKEN_DEVICE",
		"TOKEN_RACK",
		"TOKEN_ROOM",
		"TOKEN_SITE",
		"TOKEN_TENANT",
		"E",
		"GETOBJS",
		"TOKEN_CD",
		"TOKEN_CREATE",
		"TOKEN_DELETE",
		"TOKEN_GET",
		"TOKEN_LBRAC",
		"TOKEN_LS",
		"TOKEN_LSOG",
		"TOKEN_SUBDEVICE",
		"TOKEN_SUBDEVICE1",
		"TOKEN_TREE",
		"TOKEN_UPDATE",
		"BASH",
		"K",
		"NT_CREATE",
		"NT_DEL",
		"NT_GET",
		"NT_UPDATE",
		"OCCHOOSE",
		"OCCR",
		"OCDEL",
		"OCDOT",
		"OCGET",
		"OCLISYNTX",
		"OCSEL",
		"OCUPDATE",
		"Q",
		"start",
		"TOKEN_CLR",
		"TOKEN_CMDS",
		"TOKEN_COMMA",
		"TOKEN_DEREF",
		"TOKEN_DOC",
		"TOKEN_EXIT",
		"TOKEN_GREP",
		"TOKEN_LSBLDG",
		"TOKEN_LSDEV",
		"TOKEN_LSRACK",
		"TOKEN_LSROOM",
		"TOKEN_LSSITE",
		"TOKEN_LSSUBDEV",
		"TOKEN_LSSUBDEV1",
		"TOKEN_LSTEN",
		"TOKEN_OCBLDG",
		"TOKEN_OCDEV",
		"TOKEN_OCRACK",
		"TOKEN_OCROOM",
		"TOKEN_OCSITE",
		"TOKEN_OCTENANT",
		"TOKEN_PWD",
		"TOKEN_SELECT",
		"TOKEN_TEMPLATE",
		"TOKEN_VAR",
		"$default",
		"error",
		"TOKEN_BASHTYPE",
		"TOKEN_CMDFLAG",
		"TOKEN_OCSDEV",
		"TOKEN_OCSDEV1",
		"TOKEN_SEARCH",
	}

	yyTokenLiteralStrings = map[int]string{}

	yyReductions = map[int]struct{ xsym, components int }{
		0:  {0, 1},
		1:  {51, 1},
		2:  {51, 1},
		3:  {51, 1},
		4:  {37, 1},
		5:  {37, 1},
		6:  {37, 1},
		7:  {37, 1},
		8:  {38, 3},
		9:  {38, 4},
		10: {40, 2},
		11: {40, 3},
		12: {41, 3},
		13: {39, 2},
		14: {23, 1},
		15: {23, 1},
		16: {23, 1},
		17: {23, 1},
		18: {23, 1},
		19: {23, 1},
		20: {23, 1},
		21: {23, 1},
		22: {11, 1},
		23: {11, 1},
		24: {11, 0},
		25: {12, 1},
		26: {12, 1},
		27: {12, 4},
		28: {14, 4},
		29: {14, 3},
		30: {7, 1},
		31: {7, 2},
		32: {5, 3},
		33: {5, 1},
		34: {5, 4},
		35: {5, 1},
		36: {5, 2},
		37: {5, 0},
		38: {50, 2},
		39: {50, 2},
		40: {50, 2},
		41: {50, 2},
		42: {50, 2},
		43: {50, 2},
		44: {50, 2},
		45: {50, 2},
		46: {50, 2},
		47: {50, 2},
		48: {50, 2},
		49: {50, 2},
		50: {50, 3},
		51: {50, 1},
		52: {36, 1},
		53: {36, 1},
		54: {36, 1},
		55: {36, 1},
		56: {36, 1},
		57: {36, 1},
		58: {36, 2},
		59: {36, 2},
		60: {36, 2},
		61: {36, 2},
		62: {36, 2},
		63: {36, 2},
		64: {36, 2},
		65: {36, 2},
		66: {36, 2},
		67: {47, 2},
		68: {47, 1},
		69: {47, 1},
		70: {47, 1},
		71: {47, 1},
		72: {47, 1},
		73: {47, 1},
		74: {43, 5},
		75: {43, 5},
		76: {43, 5},
		77: {43, 5},
		78: {43, 7},
		79: {43, 7},
		80: {43, 7},
		81: {43, 7},
		82: {43, 7},
		83: {43, 7},
		84: {43, 7},
		85: {43, 7},
		86: {44, 2},
		87: {49, 5},
		88: {46, 2},
		89: {24, 3},
		90: {24, 1},
		91: {42, 4},
		92: {45, 6},
		93: {45, 4},
		94: {45, 4},
		95: {45, 4},
		96: {48, 1},
		97: {48, 5},
	}

	yyXErrors = map[yyXError]string{}

	yyParseTab = [213][]uint16{
		// 0
		{1: 113, 114, 4: 112, 111, 7: 141, 9: 140, 133, 15: 142, 25: 115, 107, 110, 108, 30: 116, 129, 34: 125, 109, 126, 100, 103, 106, 104, 105, 137, 44: 134, 138, 136, 102, 139, 135, 101, 99, 127, 55: 143, 132, 131, 128, 119, 122, 121, 120, 118, 123, 124, 117, 73: 130, 144},
		{98},
		{97},
		{96},
		{95},
		// 5
		{94},
		{93},
		{92},
		{91},
		{17: 300, 303, 302, 301, 299, 298, 307, 32: 304, 305},
		// 10
		{61, 113, 152, 4: 112, 111, 7: 296, 17: 300, 303, 302, 301, 299, 298, 297, 32: 304, 305},
		{1: 113, 152, 4: 112, 111, 7: 290, 61},
		{61, 113, 152, 4: 112, 111, 7: 289},
		{68, 2: 68, 68, 6: 68, 8: 68},
		{61, 113, 152, 61, 5: 288, 61, 8: 61},
		// 15
		{65, 2: 65, 65, 286, 6: 65, 8: 65},
		{2: 160, 53: 276, 75: 277, 275},
		{61, 113, 152, 4: 112, 111, 7: 274},
		{61, 113, 152, 4: 112, 111, 7: 273},
		{61, 113, 152, 4: 112, 111, 7: 272},
		// 20
		{61, 113, 152, 4: 112, 111, 7: 271},
		{61, 113, 152, 4: 112, 111, 7: 270},
		{61, 113, 152, 4: 112, 111, 7: 269},
		{61, 113, 152, 4: 112, 111, 7: 268},
		{61, 113, 152, 4: 112, 111, 7: 267},
		// 25
		{61, 113, 152, 4: 112, 111, 7: 266},
		{61, 113, 152, 4: 112, 111, 7: 265},
		{61, 113, 152, 4: 112, 111, 262, 263},
		{47},
		{46},
		// 30
		{45},
		{44},
		{43},
		{42},
		{41, 259, 25: 254, 255, 258, 256, 30: 253, 261, 34: 260, 257},
		// 35
		{17: 182, 188, 186, 184, 180, 178, 43: 176, 67: 181, 187, 185, 183, 179, 177},
		{30},
		{29},
		{28},
		{27},
		// 40
		{26},
		{25},
		{61, 113, 152, 4: 112, 111, 7: 175},
		{2: 163},
		{61, 113, 152, 4: 112, 111, 7: 153, 29: 154},
		// 45
		{29: 149},
		{2, 2: 145},
		{8: 146},
		{15: 147},
		{1: 148},
		// 50
		{1},
		{1: 150},
		{16: 151},
		{3},
		{2: 160},
		// 55
		{10},
		{1: 155, 24: 156},
		{16: 8, 54: 158},
		{16: 157},
		{7},
		// 60
		{1: 155, 24: 159},
		{16: 9},
		{62, 2: 62, 62, 161, 6: 62, 8: 62},
		{61, 113, 152, 61, 5: 162, 61, 8: 61},
		{64, 2: 64, 64, 6: 64, 8: 64},
		// 65
		{8: 164},
		{15: 165},
		{1: 168, 6: 169, 9: 167, 166, 170, 171},
		{1: 76},
		{1: 75},
		// 70
		{73, 3: 73, 8: 73},
		{72, 3: 72, 8: 72},
		{1: 172},
		{11},
		{1: 74, 9: 167, 166, 173},
		// 75
		{1: 174},
		{71, 3: 71, 8: 71},
		{12},
		{31},
		{13: 249},
		// 80
		{13: 245},
		{13: 241},
		{13: 237},
		{13: 231},
		{13: 225},
		// 85
		{13: 219},
		{13: 213},
		{13: 207},
		{13: 201},
		{13: 195},
		// 90
		{13: 189},
		{1: 113, 152, 61, 112, 111, 7: 190},
		{3: 191},
		{1: 168, 6: 169, 9: 167, 166, 170, 192},
		{3: 193},
		// 95
		{1: 168, 6: 169, 9: 167, 166, 170, 194},
		{13},
		{1: 113, 152, 61, 112, 111, 7: 196},
		{3: 197},
		{1: 168, 6: 169, 9: 167, 166, 170, 198},
		// 100
		{3: 199},
		{1: 168, 6: 169, 9: 167, 166, 170, 200},
		{14},
		{1: 113, 152, 61, 112, 111, 7: 202},
		{3: 203},
		// 105
		{1: 168, 6: 169, 9: 167, 166, 170, 204},
		{3: 205},
		{1: 168, 6: 169, 9: 167, 166, 170, 206},
		{15},
		{1: 113, 152, 61, 112, 111, 7: 208},
		// 110
		{3: 209},
		{1: 168, 6: 169, 9: 167, 166, 170, 210},
		{3: 211},
		{1: 168, 6: 169, 9: 167, 166, 170, 212},
		{16},
		// 115
		{1: 113, 152, 61, 112, 111, 7: 214},
		{3: 215},
		{1: 168, 6: 169, 9: 167, 166, 170, 216},
		{3: 217},
		{1: 168, 6: 169, 9: 167, 166, 170, 218},
		// 120
		{17},
		{1: 113, 152, 61, 112, 111, 7: 220},
		{3: 221},
		{1: 168, 6: 169, 9: 167, 166, 170, 222},
		{3: 223},
		// 125
		{1: 168, 6: 169, 9: 167, 166, 170, 224},
		{18},
		{1: 113, 152, 61, 112, 111, 7: 226},
		{3: 227},
		{1: 168, 6: 169, 9: 167, 166, 170, 228},
		// 130
		{3: 229},
		{1: 168, 6: 169, 9: 167, 166, 170, 230},
		{19},
		{1: 113, 152, 61, 112, 111, 7: 232},
		{3: 233},
		// 135
		{1: 168, 6: 169, 9: 167, 166, 170, 234},
		{3: 235},
		{1: 168, 6: 169, 9: 167, 166, 170, 236},
		{20},
		{1: 113, 152, 61, 112, 111, 7: 238},
		// 140
		{3: 239},
		{1: 168, 6: 169, 9: 167, 166, 170, 240},
		{21},
		{1: 113, 152, 61, 112, 111, 7: 242},
		{3: 243},
		// 145
		{1: 168, 6: 169, 9: 167, 166, 170, 244},
		{22},
		{1: 113, 152, 61, 112, 111, 7: 246},
		{3: 247},
		{1: 168, 6: 169, 9: 167, 166, 170, 248},
		// 150
		{23},
		{1: 113, 152, 61, 112, 111, 7: 250},
		{3: 251},
		{1: 168, 6: 169, 9: 167, 166, 170, 252},
		{24},
		// 155
		{40},
		{39},
		{38},
		{37},
		{36},
		// 160
		{35},
		{34},
		{33},
		{32},
		{50},
		// 165
		{49, 6: 264},
		{48},
		{51},
		{52},
		{53},
		// 170
		{54},
		{55},
		{56},
		{57},
		{58},
		// 175
		{59},
		{60},
		{13: 282},
		{13: 280},
		{13: 278},
		// 180
		{61, 113, 152, 4: 112, 111, 7: 279},
		{4},
		{61, 113, 152, 4: 112, 111, 7: 281},
		{5},
		{1: 283},
		// 185
		{15: 284},
		{1: 168, 6: 169, 9: 167, 166, 170, 285},
		{6},
		{61, 113, 152, 61, 5: 287, 61, 8: 61},
		{66, 2: 66, 66, 6: 66, 8: 66},
		// 190
		{67, 2: 67, 67, 6: 67, 8: 67},
		{85},
		{8: 292, 14: 291},
		{86},
		{15: 293},
		// 195
		{1: 168, 6: 169, 9: 167, 166, 170, 294},
		{69, 8: 292, 14: 295},
		{70},
		{88},
		{8: 292, 14: 306},
		// 200
		{1: 84, 84, 4: 84, 8: 84},
		{1: 83, 83, 4: 83, 8: 83},
		{1: 82, 82, 4: 82, 8: 82},
		{1: 81, 81, 4: 81, 8: 81},
		{1: 80, 80, 4: 80, 8: 80},
		// 205
		{1: 79, 79, 4: 79, 8: 79},
		{1: 78, 78, 4: 78, 8: 78},
		{1: 77, 77, 4: 77, 8: 77},
		{87},
		{1: 113, 152, 4: 112, 111, 7: 309, 292, 14: 308},
		// 210
		{90},
		{8: 292, 14: 310},
		{89},
	}
)

var yyDebug = 0

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyLexerEx interface {
	yyLexer
	Reduced(rule, state int, lval *yySymType) bool
}

func yySymName(c int) (s string) {
	x, ok := yyXLAT[c]
	if ok {
		return yySymNames[x]
	}

	if c < 0x7f {
		return __yyfmt__.Sprintf("%q", c)
	}

	return __yyfmt__.Sprintf("%d", c)
}

func yylex1(yylex yyLexer, lval *yySymType) (n int) {
	n = yylex.Lex(lval)
	if n <= 0 {
		n = yyEofCode
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("\nlex %s(%#x %d), lval: %+v\n", yySymName(n), n, n, lval)
	}
	return n
}

func yyParse(yylex yyLexer) int {
	const yyError = 78

	yyEx, _ := yylex.(yyLexerEx)
	var yyn int
	var yylval yySymType
	var yyVAL yySymType
	yyS := make([]yySymType, 200)

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yyerrok := func() {
		if yyDebug >= 2 {
			__yyfmt__.Printf("yyerrok()\n")
		}
		Errflag = 0
	}
	_ = yyerrok
	yystate := 0
	yychar := -1
	var yyxchar int
	var yyshift int
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	if yychar < 0 {
		yylval.yys = yystate
		yychar = yylex1(yylex, &yylval)
		var ok bool
		if yyxchar, ok = yyXLAT[yychar]; !ok {
			yyxchar = len(yySymNames) // > tab width
		}
	}
	if yyDebug >= 4 {
		var a []int
		for _, v := range yyS[:yyp+1] {
			a = append(a, v.yys)
		}
		__yyfmt__.Printf("state stack %v\n", a)
	}
	row := yyParseTab[yystate]
	yyn = 0
	if yyxchar < len(row) {
		if yyn = int(row[yyxchar]); yyn != 0 {
			yyn += yyTabOfs
		}
	}
	switch {
	case yyn > 0: // shift
		yychar = -1
		yyVAL = yylval
		yystate = yyn
		yyshift = yyn
		if yyDebug >= 2 {
			__yyfmt__.Printf("shift, and goto state %d\n", yystate)
		}
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	case yyn < 0: // reduce
	case yystate == 1: // accept
		if yyDebug >= 2 {
			__yyfmt__.Println("accept")
		}
		goto ret0
	}

	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			if yyDebug >= 1 {
				__yyfmt__.Printf("no action for %s in state %d\n", yySymName(yychar), yystate)
			}
			msg, ok := yyXErrors[yyXError{yystate, yyxchar}]
			if !ok {
				msg, ok = yyXErrors[yyXError{yystate, -1}]
			}
			if !ok && yyshift != 0 {
				msg, ok = yyXErrors[yyXError{yyshift, yyxchar}]
			}
			if !ok {
				msg, ok = yyXErrors[yyXError{yyshift, -1}]
			}
			if yychar > 0 {
				ls := yyTokenLiteralStrings[yychar]
				if ls == "" {
					ls = yySymName(yychar)
				}
				if ls != "" {
					switch {
					case msg == "":
						msg = __yyfmt__.Sprintf("unexpected %s", ls)
					default:
						msg = __yyfmt__.Sprintf("unexpected %s, %s", ls, msg)
					}
				}
			}
			if msg == "" {
				msg = "syntax error"
			}
			println("OGREE: Unrecognised command!")
cmd.WarningLogger.Println("Unknown Command")			/*yylex.Error(msg)*/
			Nerrs++
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				row := yyParseTab[yyS[yyp].yys]
				if yyError < len(row) {
					yyn = int(row[yyError]) + yyTabOfs
					if yyn > 0 { // hit
						if yyDebug >= 2 {
							__yyfmt__.Printf("error recovery found error shift in state %d\n", yyS[yyp].yys)
						}
						yystate = yyn /* simulate a shift of "error" */
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery failed\n")
			}
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yySymName(yychar))
			}
			if yychar == yyEofCode {
				goto ret1
			}

			yychar = -1
			goto yynewstate /* try again in the same state */
		}
	}

	r := -yyn
	x0 := yyReductions[r]
	x, n := x0.xsym, x0.components
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= n
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	exState := yystate
	yystate = int(yyParseTab[yyS[yyp].yys][x]) + yyTabOfs
	/* reduction by production r */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce using rule %v (%s), and goto state %d\n", r, yySymNames[x], yystate)
	}

	switch r {
	case 4:
		{
			println("@State start")
		}
	case 8:
		{
			cmd.PostObj(cmd.EntityStrToInt(yyS[yypt-1].s), yyS[yypt-1].s, resMap(&yyS[yypt-0].s))
		}
	case 9:
		{
			yyVAL.s = yyS[yypt-0].s
			cmd.Disp(resMap(&yyS[yypt-0].s))
			cmd.PostObj(cmd.EntityStrToInt(yyS[yypt-2].s), yyS[yypt-2].s, resMap(&yyS[yypt-0].s))
		}
	case 10:
		{
			cmd.GetObject(yyS[yypt-0].s)
		}
	case 11:
		{ /*cmd.Disp(resMap(&$4)); */
			cmd.SearchObjects(yyS[yypt-1].s, resMap(&yyS[yypt-0].s))
		}
	case 12:
		{
			yyVAL.s = yyS[yypt-0].s /*cmd.Disp(resMap(&$4));*/
			cmd.UpdateObj(yyS[yypt-1].s, resMap(&yyS[yypt-0].s))
		}
	case 13:
		{
			println("@State NT_DEL")
			cmd.DeleteObj(yyS[yypt-0].s)
		}
	case 22:
		{
			yyVAL.s = yyS[yypt-0].s
		}
	case 23:
		{
			yyVAL.s = yyS[yypt-0].s
		}
	case 24:
		{
			yyVAL.s = ""
		}
	case 25:
		{
			yyVAL.s = yyS[yypt-0].s
		}
	case 26:
		{
			x := strconv.Itoa(yyS[yypt-0].n)
			yyVAL.s = x
		}
	case 27:
		{
			yyVAL.s = yyS[yypt-3].s + yyS[yypt-2].s + yyS[yypt-1].s + yyS[yypt-0].s
		}
	case 28:
		{
			yyVAL.s = string(yyS[yypt-3].s + "=" + yyS[yypt-1].s + "=" + yyS[yypt-0].s)
			println("So we got: ", yyVAL.s)
		}
	case 29:
		{
			yyVAL.s = yyS[yypt-2].s + "=" + yyS[yypt-0].s
		}
	case 31:
		{
			yyVAL.s = "/" + yyS[yypt-0].s
		}
	case 32:
		{
			yyVAL.s = yyS[yypt-2].s + "/" + yyS[yypt-0].s
		}
	case 33:
		{
			yyVAL.s = yyS[yypt-0].s
		}
	case 34:
		{
			yyVAL.s = "../" + yyS[yypt-0].s
		}
	case 35:
		{
			yyVAL.s = yyS[yypt-0].s
		}
	case 36:
		{
			yyVAL.s = ".."
		}
	case 37:
		{
			yyVAL.s = ""
		}
	case 38:
		{
			cmd.CD(yyS[yypt-0].s)
		}
	case 39:
		{
			cmd.LS(yyS[yypt-0].s)
		}
	case 40:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 0)
		}
	case 41:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 1)
		}
	case 42:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 2)
		}
	case 43:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 3)
		}
	case 44:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 4)
		}
	case 45:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 5)
		}
	case 46:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 6)
		}
	case 47:
		{
			cmd.LSOBJECT(yyS[yypt-0].s, 7)
		}
	case 48:
		{
			cmd.Tree("", yyS[yypt-0].n)
		}
	case 49:
		{
			cmd.Tree(yyS[yypt-0].s, 0)
		}
	case 50:
		{
			cmd.Tree(yyS[yypt-1].s, yyS[yypt-0].n)
		}
	case 51:
		{
			cmd.Execute()
		}
	case 54:
		{
			cmd.LSOG()
		}
	case 55:
		{
			cmd.PWD()
		}
	case 56:
		{
			cmd.Exit()
		}
	case 57:
		{
			cmd.Help("")
		}
	case 58:
		{
			cmd.Help("ls")
		}
	case 59:
		{
			cmd.Help("cd")
		}
	case 60:
		{
			cmd.Help("create")
		}
	case 61:
		{
			cmd.Help("gt")
		}
	case 62:
		{
			cmd.Help("update")
		}
	case 63:
		{
			cmd.Help("delete")
		}
	case 64:
		{
			cmd.Help(yyS[yypt-0].s)
		}
	case 65:
		{
			cmd.Help("tree")
		}
	case 66:
		{
			cmd.Help("lsog")
		}
	case 74:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-2].s)), cmd.TENANT, map[string]interface{}{"attributes": map[string]interface{}{"color": yyS[yypt-0].s}}, rlPtr)
		}
	case 75:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-2].s)), cmd.TENANT, map[string]interface{}{"attributes": map[string]interface{}{"color": yyS[yypt-0].s}}, rlPtr)
		}
	case 76:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-2].s)), cmd.SITE, map[string]interface{}{"attributes": map[string]interface{}{"orientation": yyS[yypt-0].s}}, rlPtr)
		}
	case 77:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-2].s)), cmd.SITE, map[string]interface{}{"attributes": map[string]interface{}{"orientation": yyS[yypt-0].s}}, rlPtr)
		}
	case 78:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.BLDG, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 79:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.BLDG, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 80:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.ROOM, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 81:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.ROOM, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 82:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.RACK, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 83:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.RACK, map[string]interface{}{"attributes": map[string]interface{}{"posXY": yyS[yypt-2].s, "size": yyS[yypt-0].s}}, rlPtr)
		}
	case 84:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.DEVICE, map[string]interface{}{"attributes": map[string]interface{}{"slot": yyS[yypt-2].s, "sizeUnit": yyS[yypt-0].s}}, rlPtr)
		}
	case 85:
		{
			cmd.GetOCLIAtrributes(cmd.StrToStack(replaceOCLICurrPath(yyS[yypt-4].s)), cmd.DEVICE, map[string]interface{}{"attributes": map[string]interface{}{"slot": yyS[yypt-2].s, "sizeUnit": yyS[yypt-0].s}}, rlPtr)
		}
	case 86:
		{
			cmd.DeleteObj(replaceOCLICurrPath(yyS[yypt-0].s))
		}
	case 87:
		{
			println("Attribute Acquired")
			val := yyS[yypt-2].s + "=" + yyS[yypt-0].s
			cmd.UpdateObj(replaceOCLICurrPath(yyS[yypt-4].s), resMap(&val))
		}
	case 88:
		{
			cmd.GetObject(replaceOCLICurrPath(yyS[yypt-0].s))
		}
	case 89:
		{
			x := make([]string, 0)
			x = append(x, cmd.State.CurrPath+"/"+yyS[yypt-2].s)
			x = append(x, yyS[yypt-0].sarr...)
			yyVAL.sarr = x
		}
	case 90:
		{
			yyVAL.sarr = []string{cmd.State.CurrPath + "/" + yyS[yypt-0].s}
		}
	case 91:
		{
			cmd.State.ClipBoard = &yyS[yypt-1].sarr
			println("Selection made!")
		}
	case 92:
		{
			println("You want to assign", yyS[yypt-2].s, "with value of", yyS[yypt-0].s)
		}
	case 93:
		{
			cmd.LoadFile(yyS[yypt-0].s)
		}
	case 94:
		{
			cmd.LoadFile(yyS[yypt-0].s)
		}
	case 95:
		{
			println("So You want the value")
		}
	case 96:
		{
			cmd.ShowClipBoard()
		}
	case 97:
		{
			x := yyS[yypt-2].s + "=" + yyS[yypt-0].s
			cmd.UpdateSelection(resMap(&x))
		}

	}

	if yyEx != nil && yyEx.Reduced(r, exState, &yyVAL) {
		return -1
	}
	goto yystack /* stack new state and value */
}
