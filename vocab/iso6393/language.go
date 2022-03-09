// Code generated by go generate; DO NOT EDIT.
// This file was generated at 2022-03-09T12:35:08+01:00
// using data from https://www.wikidata.org
package iso6393

import (
	"fmt"
	"sort"

	"golang.org/x/text/language"
	"github.com/knakk/sirkulator/internal/localizer"
)

//go:generate go run gen_language.go
//go:generate go fmt language.go

// Language is a iso 639-3 language, represented by a 3-letter code.
type Language string

const (
	AAR Language = "aar"
	ABK Language = "abk"
	ACE Language = "ace"
	ACH Language = "ach"
	ADA Language = "ada"
	ADY Language = "ady"
	AFH Language = "afh"
	AFR Language = "afr"
	AIN Language = "ain"
	AKA Language = "aka"
	AKK Language = "akk"
	ALB Language = "sqi"
	ALE Language = "ale"
	ALT Language = "alt"
	AMH Language = "amh"
	ANG Language = "ang"
	ANP Language = "anp"
	ARA Language = "ara"
	ARC Language = "arc"
	ARG Language = "arg"
	ARM Language = "hye"
	ARN Language = "arn"
	ARP Language = "arp"
	ARW Language = "arw"
	ASM Language = "asm"
	AST Language = "ast"
	AVA Language = "ava"
	AVE Language = "ave"
	AWA Language = "awa"
	AYM Language = "aym"
	AZE Language = "aze"
	BAK Language = "bak"
	BAL Language = "bal"
	BAM Language = "bam"
	BAN Language = "ban"
	BAQ Language = "eus"
	BAS Language = "bas"
	BEJ Language = "bej"
	BEL Language = "bel"
	BEM Language = "bem"
	BEN Language = "ben"
	BHO Language = "bho"
	BIK Language = "bik"
	BIN Language = "bin"
	BIS Language = "bis"
	BLA Language = "bla"
	BOS Language = "bos"
	BRA Language = "bra"
	BRE Language = "bre"
	BUA Language = "bua"
	BUG Language = "bug"
	BUL Language = "bul"
	BUR Language = "mya"
	BYN Language = "byn"
	CAD Language = "cad"
	CAR Language = "car"
	CAT Language = "cat"
	CEB Language = "ceb"
	CHA Language = "cha"
	CHB Language = "chb"
	CHE Language = "che"
	CHG Language = "chg"
	CHI Language = "zho"
	CHK Language = "chk"
	CHM Language = "chm"
	CHN Language = "chn"
	CHO Language = "cho"
	CHP Language = "chp"
	CHR Language = "chr"
	CHU Language = "chu"
	CHV Language = "chv"
	CHY Language = "chy"
	CNR Language = "cnr"
	COP Language = "cop"
	COR Language = "cor"
	COS Language = "cos"
	CRE Language = "cre"
	CRH Language = "crh"
	CSB Language = "csb"
	CZE Language = "ces"
	DAK Language = "dak"
	DAN Language = "dan"
	DAR Language = "dar"
	DEL Language = "del"
	DEN Language = "den"
	DGR Language = "dgr"
	DIN Language = "din"
	DIV Language = "div"
	DOI Language = "doi"
	DSB Language = "dsb"
	DUA Language = "dua"
	DUM Language = "dum"
	DUT Language = "nld"
	DYU Language = "dyu"
	DZO Language = "dzo"
	EFI Language = "efi"
	EGY Language = "egy"
	EKA Language = "eka"
	ELX Language = "elx"
	ENG Language = "eng"
	ENM Language = "enm"
	EPO Language = "epo"
	EST Language = "est"
	EWE Language = "ewe"
	EWO Language = "ewo"
	FAN Language = "fan"
	FAO Language = "fao"
	FAT Language = "fat"
	FIJ Language = "fij"
	FIL Language = "fil"
	FIN Language = "fin"
	FON Language = "fon"
	FRE Language = "fra"
	FRM Language = "frm"
	FRO Language = "fro"
	FRR Language = "frr"
	FRS Language = "frs"
	FRY Language = "fry"
	FUL Language = "ful"
	FUR Language = "fur"
	GAA Language = "gaa"
	GAY Language = "gay"
	GBA Language = "gba"
	GEO Language = "kat"
	GER Language = "deu"
	GEZ Language = "gez"
	GIL Language = "gil"
	GLA Language = "gla"
	GLE Language = "gle"
	GLG Language = "glg"
	GLV Language = "glv"
	GMH Language = "gmh"
	GOH Language = "goh"
	GON Language = "gon"
	GOR Language = "gor"
	GOT Language = "got"
	GRB Language = "grb"
	GRC Language = "grc"
	GRE Language = "ell"
	GRN Language = "grn"
	GSW Language = "gsw"
	GUJ Language = "guj"
	GWI Language = "gwi"
	HAI Language = "hai"
	HAT Language = "hat"
	HAU Language = "hau"
	HAW Language = "haw"
	HEB Language = "heb"
	HER Language = "her"
	HIL Language = "hil"
	HIN Language = "hin"
	HIT Language = "hit"
	HMN Language = "hmn"
	HMO Language = "hmo"
	HRV Language = "hrv"
	HSB Language = "hsb"
	HUN Language = "hun"
	HUP Language = "hup"
	IBA Language = "iba"
	IBO Language = "ibo"
	ICE Language = "isl"
	IDO Language = "ido"
	III Language = "iii"
	IKU Language = "iku"
	ILE Language = "ile"
	ILO Language = "ilo"
	INA Language = "ina"
	IND Language = "ind"
	INH Language = "inh"
	IPK Language = "ipk"
	ITA Language = "ita"
	JAV Language = "jav"
	JBO Language = "jbo"
	JPN Language = "jpn"
	JPR Language = "jpr"
	JRB Language = "jrb"
	KAA Language = "kaa"
	KAB Language = "kab"
	KAC Language = "kac"
	KAL Language = "kal"
	KAM Language = "kam"
	KAN Language = "kan"
	KAS Language = "kas"
	KAU Language = "kau"
	KAW Language = "kaw"
	KAZ Language = "kaz"
	KBD Language = "kbd"
	KHA Language = "kha"
	KHM Language = "khm"
	KHO Language = "kho"
	KIK Language = "kik"
	KIN Language = "kin"
	KIR Language = "kir"
	KMB Language = "kmb"
	KOK Language = "kok"
	KOM Language = "kom"
	KON Language = "kon"
	KOR Language = "kor"
	KOS Language = "kos"
	KPE Language = "kpe"
	KRC Language = "krc"
	KRL Language = "krl"
	KRU Language = "kru"
	KUA Language = "kua"
	KUM Language = "kum"
	KUR Language = "kur"
	KUT Language = "kut"
	LAD Language = "lad"
	LAH Language = "lah"
	LAM Language = "lam"
	LAO Language = "lao"
	LAT Language = "lat"
	LAV Language = "lav"
	LEZ Language = "lez"
	LIM Language = "lim"
	LIN Language = "lin"
	LIT Language = "lit"
	LOL Language = "lol"
	LOZ Language = "loz"
	LTZ Language = "ltz"
	LUA Language = "lua"
	LUB Language = "lub"
	LUG Language = "lug"
	LUI Language = "lui"
	LUN Language = "lun"
	LUO Language = "luo"
	LUS Language = "lus"
	MAC Language = "mkd"
	MAD Language = "mad"
	MAG Language = "mag"
	MAH Language = "mah"
	MAI Language = "mai"
	MAK Language = "mak"
	MAL Language = "mal"
	MAN Language = "man"
	MAO Language = "mri"
	MAR Language = "mar"
	MAS Language = "mas"
	MAY Language = "msa"
	MDF Language = "mdf"
	MDR Language = "mdr"
	MEN Language = "men"
	MGA Language = "mga"
	MIC Language = "mic"
	MIN Language = "min"
	MLG Language = "mlg"
	MLT Language = "mlt"
	MNC Language = "mnc"
	MNI Language = "mni"
	MOH Language = "moh"
	MON Language = "mon"
	MOS Language = "mos"
	MUL Language = "mul"
	MUS Language = "mus"
	MWL Language = "mwl"
	MWR Language = "mwr"
	MYV Language = "myv"
	NAI Language = "aqp"
	NAP Language = "nap"
	NAU Language = "nau"
	NAV Language = "nav"
	NBL Language = "nbl"
	NDE Language = "nde"
	NDO Language = "ndo"
	NDS Language = "nds"
	NEP Language = "nep"
	NEW Language = "new"
	NIA Language = "nia"
	NIU Language = "niu"
	NNO Language = "nno"
	NOB Language = "nob"
	NOG Language = "nog"
	NON Language = "non"
	NOR Language = "nor"
	NQO Language = "nqo"
	NSO Language = "nso"
	NWC Language = "nwc"
	NYA Language = "nya"
	NYM Language = "nym"
	NYN Language = "nyn"
	NYO Language = "nyo"
	NZI Language = "nzi"
	OCI Language = "oci"
	OJI Language = "oji"
	ORI Language = "ory"
	ORM Language = "orm"
	OSA Language = "osa"
	OSS Language = "oss"
	OTA Language = "ota"
	PAG Language = "pag"
	PAL Language = "pal"
	PAM Language = "pam"
	PAN Language = "pan"
	PAP Language = "pap"
	PAU Language = "pau"
	PEO Language = "peo"
	PER Language = "fas"
	PHN Language = "phn"
	PLI Language = "pli"
	POL Language = "pol"
	PON Language = "pon"
	POR Language = "por"
	PRO Language = "pro"
	PUS Language = "pus"
	QUE Language = "que"
	RAJ Language = "raj"
	RAP Language = "rap"
	RAR Language = "rar"
	ROH Language = "roh"
	ROM Language = "rom"
	RUM Language = "ron"
	RUN Language = "run"
	RUP Language = "rup"
	RUS Language = "rus"
	SAD Language = "sad"
	SAG Language = "sag"
	SAH Language = "sah"
	SAM Language = "sam"
	SAN Language = "san"
	SAS Language = "sas"
	SAT Language = "sat"
	SCN Language = "scn"
	SCO Language = "sco"
	SEL Language = "sel"
	SGA Language = "sga"
	SHN Language = "shn"
	SID Language = "sid"
	SIN Language = "sin"
	SLO Language = "slk"
	SLV Language = "slv"
	SMA Language = "sma"
	SME Language = "sme"
	SMJ Language = "smj"
	SMN Language = "smn"
	SMO Language = "smo"
	SMS Language = "sms"
	SNA Language = "sna"
	SND Language = "snd"
	SNK Language = "snk"
	SOG Language = "sog"
	SOM Language = "som"
	SOT Language = "sot"
	SPA Language = "spa"
	SRD Language = "srd"
	SRN Language = "srn"
	SRP Language = "srp"
	SRR Language = "srr"
	SSW Language = "ssw"
	SUK Language = "suk"
	SUN Language = "sun"
	SUS Language = "sus"
	SUX Language = "sux"
	SWA Language = "swa"
	SWE Language = "swe"
	SYC Language = "syc"
	TAH Language = "tah"
	TAM Language = "tam"
	TAT Language = "tat"
	TEL Language = "tel"
	TEM Language = "tem"
	TER Language = "ter"
	TET Language = "tet"
	TGK Language = "tgk"
	TGL Language = "tgl"
	THA Language = "tha"
	TIB Language = "bod"
	TIG Language = "tig"
	TIR Language = "tir"
	TIV Language = "tiv"
	TKL Language = "tkl"
	TLH Language = "tlh"
	TLI Language = "tli"
	TMH Language = "tmh"
	TOG Language = "tog"
	TON Language = "ton"
	TPI Language = "tpi"
	TSI Language = "tsi"
	TSN Language = "tsn"
	TSO Language = "tso"
	TUK Language = "tuk"
	TUM Language = "tum"
	TUR Language = "tur"
	TVL Language = "tvl"
	TWI Language = "twi"
	TYV Language = "tyv"
	UDM Language = "udm"
	UGA Language = "uga"
	UIG Language = "uig"
	UKR Language = "ukr"
	UMB Language = "umb"
	UND Language = "und"
	URD Language = "urd"
	UZB Language = "uzb"
	VAI Language = "vai"
	VEN Language = "ven"
	VIE Language = "vie"
	VOL Language = "vol"
	VOT Language = "vot"
	WAL Language = "wal"
	WAR Language = "war"
	WAS Language = "was"
	WEL Language = "cym"
	WLN Language = "wln"
	WOL Language = "wol"
	XAL Language = "xal"
	XHO Language = "xho"
	YAO Language = "yao"
	YAP Language = "yap"
	YID Language = "yid"
	YOR Language = "yor"
	ZAP Language = "zap"
	ZBL Language = "zbl"
	ZEN Language = "zen"
	ZHA Language = "zha"
	ZUL Language = "zul"
	ZUN Language = "zun"
	ZXX Language = "zxx"
	ZZA Language = "zza"
)

var allLanguages = []Language{ 
	AAR,ABK,ACE,ACH,ADA,ADY,AFH,AFR,AIN,AKA,AKK,ALB,ALE,ALT,AMH,ANG,ANP,ARA,ARC,ARG, 
	ARM,ARN,ARP,ARW,ASM,AST,AVA,AVE,AWA,AYM,AZE,BAK,BAL,BAM,BAN,BAQ,BAS,BEJ,BEL,BEM, 
	BEN,BHO,BIK,BIN,BIS,BLA,BOS,BRA,BRE,BUA,BUG,BUL,BUR,BYN,CAD,CAR,CAT,CEB,CHA,CHB, 
	CHE,CHG,CHI,CHK,CHM,CHN,CHO,CHP,CHR,CHU,CHV,CHY,CNR,COP,COR,COS,CRE,CRH,CSB,CZE, 
	DAK,DAN,DAR,DEL,DEN,DGR,DIN,DIV,DOI,DSB,DUA,DUM,DUT,DYU,DZO,EFI,EGY,EKA,ELX,ENG, 
	ENM,EPO,EST,EWE,EWO,FAN,FAO,FAT,FIJ,FIL,FIN,FON,FRE,FRM,FRO,FRR,FRS,FRY,FUL,FUR, 
	GAA,GAY,GBA,GEO,GER,GEZ,GIL,GLA,GLE,GLG,GLV,GMH,GOH,GON,GOR,GOT,GRB,GRC,GRE,GRN, 
	GSW,GUJ,GWI,HAI,HAT,HAU,HAW,HEB,HER,HIL,HIN,HIT,HMN,HMO,HRV,HSB,HUN,HUP,IBA,IBO, 
	ICE,IDO,III,IKU,ILE,ILO,INA,IND,INH,IPK,ITA,JAV,JBO,JPN,JPR,JRB,KAA,KAB,KAC,KAL, 
	KAM,KAN,KAS,KAU,KAW,KAZ,KBD,KHA,KHM,KHO,KIK,KIN,KIR,KMB,KOK,KOM,KON,KOR,KOS,KPE, 
	KRC,KRL,KRU,KUA,KUM,KUR,KUT,LAD,LAH,LAM,LAO,LAT,LAV,LEZ,LIM,LIN,LIT,LOL,LOZ,LTZ, 
	LUA,LUB,LUG,LUI,LUN,LUO,LUS,MAC,MAD,MAG,MAH,MAI,MAK,MAL,MAN,MAO,MAR,MAS,MAY,MDF, 
	MDR,MEN,MGA,MIC,MIN,MLG,MLT,MNC,MNI,MOH,MON,MOS,MUL,MUS,MWL,MWR,MYV,NAI,NAP,NAU, 
	NAV,NBL,NDE,NDO,NDS,NEP,NEW,NIA,NIU,NNO,NOB,NOG,NON,NOR,NQO,NSO,NWC,NYA,NYM,NYN, 
	NYO,NZI,OCI,OJI,ORI,ORM,OSA,OSS,OTA,PAG,PAL,PAM,PAN,PAP,PAU,PEO,PER,PHN,PLI,POL, 
	PON,POR,PRO,PUS,QUE,RAJ,RAP,RAR,ROH,ROM,RUM,RUN,RUP,RUS,SAD,SAG,SAH,SAM,SAN,SAS, 
	SAT,SCN,SCO,SEL,SGA,SHN,SID,SIN,SLO,SLV,SMA,SME,SMJ,SMN,SMO,SMS,SNA,SND,SNK,SOG, 
	SOM,SOT,SPA,SRD,SRN,SRP,SRR,SSW,SUK,SUN,SUS,SUX,SWA,SWE,SYC,TAH,TAM,TAT,TEL,TEM, 
	TER,TET,TGK,TGL,THA,TIB,TIG,TIR,TIV,TKL,TLH,TLI,TMH,TOG,TON,TPI,TSI,TSN,TSO,TUK, 
	TUM,TUR,TVL,TWI,TYV,UDM,UGA,UIG,UKR,UMB,UND,URD,UZB,VAI,VEN,VIE,VOL,VOT,WAL,WAR, 
	WAS,WEL,WLN,WOL,XAL,XHO,YAO,YAP,YID,YOR,ZAP,ZBL,ZEN,ZHA,ZUL,ZUN,ZXX,ZZA,
}

var languageLabels = map[Language][2]string{
	AAR: {"Afar", "afar"},
	ABK: {"Abkhaz", "abkhasisk"},
	ACE: {"Acehnese", ""},
	ACH: {"Acholi", "Acholi"},
	ADA: {"Dangme", ""},
	ADY: {"Adyghe", "adygeisk"},
	AFH: {"Afrihili", "Afrihili"},
	AFR: {"Afrikaans", "afrikaans"},
	AIN: {"Ainu", "ainu"},
	AKA: {"Akan", "akan"},
	AKK: {"Akkadian", "Akkadisk"},
	ALB: {"Albanian", "albansk"},
	ALE: {"Aleut", "aleutisk"},
	ALT: {"Altai", ""},
	AMH: {"Amharic", "amharisk"},
	ANG: {"Old English", "gammelengelsk"},
	ANP: {"Angika", ""},
	ARA: {"Arabic", "arabisk"},
	ARC: {"Aramaic", "arameisk"},
	ARG: {"Aragonese", "aragonesisk"},
	ARM: {"Armenian", "armensk"},
	ARN: {"Mapudungun", "Mapudungun"},
	ARP: {"Arapaho", ""},
	ARW: {"Arawak", "arawak"},
	ASM: {"Assamese", "assamesisk"},
	AST: {"Asturian", "asturiansk"},
	AVA: {"Avaric", "avarisk"},
	AVE: {"Avestan", "Avestisk"},
	AWA: {"Awadhi", "Awadhi"},
	AYM: {"Aymara", "aymara"},
	AZE: {"Azerbaijani", "aserbajdsjansk"},
	BAK: {"Bashkir", "basjkirsk"},
	BAL: {"Baluchi", "Balutsji"},
	BAM: {"Bambara", "bambara"},
	BAN: {"Balinese", "balinesisk"},
	BAQ: {"Basque", "baskisk"},
	BAS: {"Basaa", ""},
	BEJ: {"Beja", "Beja"},
	BEL: {"Belarusian", "hviterussisk"},
	BEM: {"Bemba", "Bemba"},
	BEN: {"Bengali", "bengali"},
	BHO: {"Bhojpuri", "Bhojpuri"},
	BIK: {"Bikol", "Bikol"},
	BIN: {"Edo", "Edo"},
	BIS: {"Bislama", "bislama"},
	BLA: {"Blackfoot", "blackfoot"},
	BOS: {"Bosnian", "bosnisk"},
	BRA: {"Braj Bhasha", ""},
	BRE: {"Breton", "bretonsk"},
	BUA: {"Buryat", "burjatisk"},
	BUG: {"Buginese", "Buginesisk"},
	BUL: {"Bulgarian", "bulgarsk"},
	BUR: {"Burmese", "burmesisk"},
	BYN: {"Blin", "Bilen"},
	CAD: {"Caddo", ""},
	CAR: {"Carib", ""},
	CAT: {"Catalan", "katalansk"},
	CEB: {"Cebuano", "cebuano"},
	CHA: {"Chamorro", "chamorro"},
	CHB: {"Chibcha", "muisca"},
	CHE: {"Chechen", "tsjetsjensk"},
	CHG: {"Chagatai", "Chagatai"},
	CHI: {"Chinese", "kinesisk"},
	CHK: {"Chuukese", ""},
	CHM: {"Mari", "mariske språk"},
	CHN: {"Chinook Jargon", "Chinook jargon"},
	CHO: {"Choctaw", "choctaw"},
	CHP: {"Chipewyan", "Chipewyan"},
	CHR: {"Cherokee", "cherokesisk"},
	CHU: {"Church Slavonic", "kirkeslavisk"},
	CHV: {"Chuvash", "tsjuvasjisk"},
	CHY: {"Cheyenne", "Cheyenne"},
	CNR: {"Montenegrin", "montenegrinsk"},
	COP: {"Coptic", "koptisk"},
	COR: {"Cornish", "kornisk"},
	COS: {"Corsican", "korsikansk"},
	CRE: {"Cree", "cree"},
	CRH: {"Crimean Tatar", "krimtatarisk"},
	CSB: {"Kashubian", "kasjubisk"},
	CZE: {"Czech", "tsjekkisk"},
	DAK: {"Dakota", ""},
	DAN: {"Danish", "dansk"},
	DAR: {"Dargwa", "darginsk"},
	DEL: {"Delaware", ""},
	DEN: {"Slavey", "Slavey"},
	DGR: {"Dogrib", "Dogrib"},
	DIN: {"Dinka", ""},
	DIV: {"Dhivehi", "dhivehi"},
	DOI: {"Dogri–Kangri", "Dogri-kangri"},
	DSB: {"Lower Sorbian", ""},
	DUA: {"Duala", ""},
	DUM: {"Middle Dutch", ""},
	DUT: {"Dutch", "nederlandsk"},
	DYU: {"Dioula", ""},
	DZO: {"Dzongkha", "dzongkha"},
	EFI: {"Efik", "Efik"},
	EGY: {"Egyptian", "egyptisk"},
	EKA: {"Ekajuk", ""},
	ELX: {"Elamite", "Elamittisk"},
	ENG: {"English", "engelsk"},
	ENM: {"Middle English", "mellomengelsk"},
	EPO: {"Esperanto", "esperanto"},
	EST: {"Estonian", "estisk"},
	EWE: {"Ewe", "ewe"},
	EWO: {"Ewondo", ""},
	FAN: {"Fang", "Fang"},
	FAO: {"Faroese", "færøysk"},
	FAT: {"Fante", "Fanti"},
	FIJ: {"Fijian", "fijiansk"},
	FIL: {"Filipino", "filippinsk"},
	FIN: {"Finnish", "finsk"},
	FON: {"Fon", "fon"},
	FRE: {"French", "fransk"},
	FRM: {"Middle French", "mellomfransk"},
	FRO: {"Old French", "gammelfransk"},
	FRR: {"North Frisian", "nordfrisisk"},
	FRS: {"East Frisian Low Saxon", ""},
	FRY: {"West Frisian", "vestfrisisk"},
	FUL: {"Fula", "Fulfulde"},
	FUR: {"Friulian", "friulisk"},
	GAA: {"Ga", ""},
	GAY: {"Gayo", ""},
	GBA: {"Gbaya", ""},
	GEO: {"Georgian", "georgisk"},
	GER: {"German", "tysk"},
	GEZ: {"Ge'ez", "ge'ez"},
	GIL: {"Gilbertese", "Kiribatisk"},
	GLA: {"Scottish Gaelic", "skotsk-gælisk"},
	GLE: {"Irish", "irsk"},
	GLG: {"Galician", "galisisk"},
	GLV: {"Manx", "mansk"},
	GMH: {"Middle High German", "middelhøytysk"},
	GOH: {"Old High German", "gammelhøytysk"},
	GON: {"Gondi", "Gondi"},
	GOR: {"Gorontalo", "Gorontalo"},
	GOT: {"Gothic", "gotisk"},
	GRB: {"Grebo", ""},
	GRC: {"Ancient Greek", "gammelgresk"},
	GRE: {"Modern Greek", "nygresk"},
	GRN: {"Guarani", "guaraní"},
	GSW: {"Alemannic", "alemannisk"},
	GUJ: {"Gujarati", "gujarati"},
	GWI: {"Gwich’in", "Gwich'in"},
	HAI: {"Haida", "Haida"},
	HAT: {"Haitian Creole", "haitisk"},
	HAU: {"Hausa", "hausa"},
	HAW: {"Hawaiian", "hawaiisk"},
	HEB: {"Hebrew", "hebraisk"},
	HER: {"Herero", "herero"},
	HIL: {"Hiligaynon", ""},
	HIN: {"Hindi", "hindi"},
	HIT: {"Hittite", "hettittisk"},
	HMN: {"Hmongic languages", ""},
	HMO: {"Hiri Motu", "hiri motu"},
	HRV: {"Croatian", "kroatisk"},
	HSB: {"Upper Sorbian", ""},
	HUN: {"Hungarian", "ungarsk"},
	HUP: {"Hupa", "Hupa"},
	IBA: {"Iban", ""},
	IBO: {"Igbo", "ibo"},
	ICE: {"Icelandic", "islandsk"},
	IDO: {"Ido", "ido"},
	III: {"Nuosu", ""},
	IKU: {"Inuktitut", "inuktitut"},
	ILE: {"Interlingue", "Interlingue"},
	ILO: {"Ilocano", "ilokano"},
	INA: {"Interlingua", "Interlingua"},
	IND: {"Indonesian", "indonesisk"},
	INH: {"Ingush", "ingusjisk"},
	IPK: {"Inupiaq", "Inupiak"},
	ITA: {"Italian", "italiensk"},
	JAV: {"Javanese", "javanesisk"},
	JBO: {"Lojban", "Lojban"},
	JPN: {"Japanese", "japansk"},
	JPR: {"Judæo-Persian", ""},
	JRB: {"Judeo-Arabic", "jødearabiske språk"},
	KAA: {"Karakalpak", ""},
	KAB: {"Kabyle", "Kabylsk"},
	KAC: {"Jingpho", ""},
	KAL: {"Greenlandic", "grønlandsk"},
	KAM: {"Kamba", ""},
	KAN: {"Kannada", "kannada"},
	KAS: {"Kashmiri", "kasjmiri"},
	KAU: {"Kanuri", ""},
	KAW: {"Kawi", ""},
	KAZ: {"Kazakh", "kasakhisk"},
	KBD: {"Kabardian", "kabardisk"},
	KHA: {"Khasi", ""},
	KHM: {"Khmer", "khmer"},
	KHO: {"Khotanese", "Khotanesisk"},
	KIK: {"Gikuyu", "kikuyu"},
	KIN: {"Kinyarwanda", "kinyarwanda"},
	KIR: {"Kyrgyz", "kirgisisk"},
	KMB: {"Kimbundu", ""},
	KOK: {"Konkani", "konkani"},
	KOM: {"Komi", "syrjensk"},
	KON: {"Kongo", "Kongo"},
	KOR: {"Korean", "koreansk"},
	KOS: {"Kosraean", ""},
	KPE: {"Kpelle", ""},
	KRC: {"Karachay-Balkar", "karatsjajbalkarsk"},
	KRL: {"Karelian", "karelsk"},
	KRU: {"Kurukh", ""},
	KUA: {"Kwanyama", ""},
	KUM: {"Kumyk", "kumykisk"},
	KUR: {"Kurdish", "kurdisk"},
	KUT: {"Kutenai", ""},
	LAD: {"Judaeo-Spanish", "jødespansk"},
	LAH: {"Lahnda languages", "lahnda"},
	LAM: {"Lamba", ""},
	LAO: {"Lao", "laotisk"},
	LAT: {"Latin", "latin"},
	LAV: {"Latvian", "latvisk"},
	LEZ: {"Lezgian", "lezgisk"},
	LIM: {"Limburgish", "limburgsk"},
	LIN: {"Lingala", "lingala"},
	LIT: {"Lithuanian", "litauisk"},
	LOL: {"Mongo", ""},
	LOZ: {"Lozi", "Lozi"},
	LTZ: {"Luxembourgish", "luxembourgsk"},
	LUA: {"Luba-Kasai", "Luba"},
	LUB: {"Luba-Katanga", ""},
	LUG: {"Luganda", "Luganda"},
	LUI: {"Luiseño", ""},
	LUN: {"Lunda", ""},
	LUO: {"Dholuo", "Luo"},
	LUS: {"Mizo", ""},
	MAC: {"Macedonian", "makedonsk"},
	MAD: {"Madurese", ""},
	MAG: {"Magahi", "Magahi"},
	MAH: {"Marshallese", "marshallesisk"},
	MAI: {"Maithili", "Maithili"},
	MAK: {"Makassarese", ""},
	MAL: {"Malayalam", "malayalam"},
	MAN: {"Manding languages", ""},
	MAO: {"Māori", "maorisk"},
	MAR: {"Marathi", "marathi"},
	MAS: {"Maasai", "Masai"},
	MAY: {"Malay", "malayisk"},
	MDF: {"Moksha", "moksja"},
	MDR: {"Mandar", ""},
	MEN: {"Mende", "Mende"},
	MGA: {"Middle Irish", "mellomirsk"},
	MIC: {"Mi'kmaq", ""},
	MIN: {"Minangkabau", "Minangkabau"},
	MLG: {"Malagasy", "gassisk"},
	MLT: {"Maltese", "maltesisk"},
	MNC: {"Manchu", "Mandsjuisk"},
	MNI: {"Meitei", "Meitei-lon"},
	MOH: {"Mohawk", ""},
	MON: {"Mongolian", "mongolsk"},
	MOS: {"Mossi", "mòoré"},
	MUL: {"multiple languages", "flere språk"},
	MUS: {"Muscogee", ""},
	MWL: {"Mirandese", "mirandesisk"},
	MWR: {"Marwari", "marwari"},
	MYV: {"Erzya", "erzia"},
	NAI: {"Atakapa", ""},
	NAP: {"Neapolitan", "napolitansk"},
	NAU: {"Nauruan", "naurisk"},
	NAV: {"Navajo", "navajo"},
	NBL: {"Southern Ndebele", "sørndebele"},
	NDE: {"Northern Ndebele", "nordndebele"},
	NDO: {"Ndonga", ""},
	NDS: {"Low German", "nedertysk"},
	NEP: {"Nepali", "nepali"},
	NEW: {"Newar", "Nepal bhasa"},
	NIA: {"Nias", ""},
	NIU: {"Niuean", "Niuisk"},
	NNO: {"Nynorsk", "nynorsk"},
	NOB: {"Bokmål", "bokmål"},
	NOG: {"Nogai", "nogaisk"},
	NON: {"Old Norse", "norrønt"},
	NOR: {"Norwegian", "norsk"},
	NQO: {"N'Ko", ""},
	NSO: {"Northern Sotho", "nordsotho"},
	NWC: {"Classical Newari", ""},
	NYA: {"Chewa", "chewa"},
	NYM: {"Nyamwezi", "Nyamwezi"},
	NYN: {"Runyankole", ""},
	NYO: {"Nyoro", ""},
	NZI: {"Nzema", ""},
	OCI: {"Occitan", "oksitansk"},
	OJI: {"Ojibwe", "Ojibwa"},
	ORI: {"Odia", "oriya"},
	ORM: {"Oromo", "oromo"},
	OSA: {"Osage", "Osage"},
	OSS: {"Ossetian", "ossetisk"},
	OTA: {"Ottoman Turkish", "osmantyrkisk"},
	PAG: {"Pangasinan", ""},
	PAL: {"Middle Persian", "Middelpersisk"},
	PAM: {"Kapampangan", "Pampangansk"},
	PAN: {"Punjabi", "panjabi"},
	PAP: {"Papiamento", "papiamento"},
	PAU: {"Palauan", "Palauisk"},
	PEO: {"Old Persian", "Gammelpersisk"},
	PER: {"Persian", "persisk"},
	PHN: {"Phoenician", "fønikisk"},
	PLI: {"Pali", "pali"},
	POL: {"Polish", "polsk"},
	PON: {"Pohnpeian", "Ponapisk"},
	POR: {"Portuguese", "portugisisk"},
	PRO: {"Old Occitan", ""},
	PUS: {"Pashto", "pashto"},
	QUE: {"Quechua", "quechua"},
	RAJ: {"Rajasthani", "Rajasthani"},
	RAP: {"Rapa Nui", "Rapanui"},
	RAR: {"Cook Islands Maori", "rarotongesisk"},
	ROH: {"Romansh", "retoromansk"},
	ROM: {"Romani", "romanes"},
	RUM: {"Romanian", "rumensk"},
	RUN: {"Kirundi", "kirundi"},
	RUP: {"Aromanian", "arumensk"},
	RUS: {"Russian", "russisk"},
	SAD: {"Sandawe", ""},
	SAG: {"Sango", "sango"},
	SAH: {"Sakha", "jakutisk"},
	SAM: {"Samaritan Aramaic", ""},
	SAN: {"Sanskrit", "sanskrit"},
	SAS: {"Sasak", ""},
	SAT: {"Santali", "Santali"},
	SCN: {"Sicilian", "siciliansk"},
	SCO: {"Scots", "skotsk"},
	SEL: {"Selkup", "selkupisk"},
	SGA: {"Old Irish", "gammelirsk"},
	SHN: {"Shan", ""},
	SID: {"Sidamo", ""},
	SIN: {"Sinhala", "singalesisk"},
	SLO: {"Slovak", "slovakisk"},
	SLV: {"Slovene", "slovensk"},
	SMA: {"Southern Sami", "sørsamisk"},
	SME: {"Northern Sami", "nordsamisk"},
	SMJ: {"Lule Sami", "lulesamisk"},
	SMN: {"Inari Sami", "enaresamisk"},
	SMO: {"Samoan", "samoansk"},
	SMS: {"Skolt Sami", "skoltesamisk"},
	SNA: {"Shona", "shona"},
	SND: {"Sindhi", "sindhi"},
	SNK: {"Soninke", ""},
	SOG: {"Sogdian", ""},
	SOM: {"Somali", "somali"},
	SOT: {"Sesotho", "sotho"},
	SPA: {"Spanish", "spansk"},
	SRD: {"Sardinian", "sardisk"},
	SRN: {"Sranan Tongo", "sranan"},
	SRP: {"Serbian", "serbisk"},
	SRR: {"Serer", ""},
	SSW: {"Swazi", "swazi"},
	SUK: {"Sukuma", "Sukuma"},
	SUN: {"Sundanese", "sundanesisk"},
	SUS: {"Susu", ""},
	SUX: {"Sumerian", "sumerisk"},
	SWA: {"Swahili", "swahili"},
	SWE: {"Swedish", "svensk"},
	SYC: {"Syriac", "Gammelsyrisk"},
	TAH: {"Tahitian", "tahitisk"},
	TAM: {"Tamil", "tamilsk"},
	TAT: {"Tatar", "tatarisk"},
	TEL: {"Telugu", "telugu"},
	TEM: {"Temne", ""},
	TER: {"Terêna", ""},
	TET: {"Tetum", "tetum"},
	TGK: {"Tajik", "tadsjikisk"},
	TGL: {"Tagalog", "tagalog"},
	THA: {"Thai", "thai"},
	TIB: {"Tibetan", "tibetansk"},
	TIG: {"Tigre", "Tigré"},
	TIR: {"Tigrinya", "tigrinja"},
	TIV: {"Tiv", ""},
	TKL: {"Tokelauan", ""},
	TLH: {"Klingon", "klingon"},
	TLI: {"Tlingit", "Tlingit"},
	TMH: {"Tuareg", "Tuareg"},
	TOG: {"Tonga", ""},
	TON: {"Tongan", "Tongansk"},
	TPI: {"Tok Pisin", "tok pisin"},
	TSI: {"Tsimshian", ""},
	TSN: {"Tswana", "setswana"},
	TSO: {"Tsonga", "tsonga"},
	TUK: {"Turkmen", "turkmensk"},
	TUM: {"Tumbuka", ""},
	TUR: {"Turkish", "tyrkisk"},
	TVL: {"Tuvaluan", "tuvalsk"},
	TWI: {"Twi", "twi"},
	TYV: {"Tuvan", "tuvinsk"},
	UDM: {"Udmurt", "udmurtisk"},
	UGA: {"Ugaritic", "Ugarittisk"},
	UIG: {"Uyghur", "uigurisk"},
	UKR: {"Ukrainian", "ukrainsk"},
	UMB: {"Umbundu", ""},
	UND: {"undetermined language", "ubestemt språk"},
	URD: {"Urdu", "urdu"},
	UZB: {"Uzbek", "usbekisk"},
	VAI: {"Vai", ""},
	VEN: {"Venda", "venda"},
	VIE: {"Vietnamese", "vietnamesisk"},
	VOL: {"Volapük", "Volapük"},
	VOT: {"Votic", "votisk"},
	WAL: {"Wolaytta", ""},
	WAR: {"Waray", "waray-waray"},
	WAS: {"Washo", ""},
	WEL: {"Welsh", "walisisk"},
	WLN: {"Walloon", "vallonsk"},
	WOL: {"Wolof", "wolof"},
	XAL: {"Oirat", "Oiratisk"},
	XHO: {"Xhosa", "xhosa"},
	YAO: {"Yao", "Yao"},
	YAP: {"Yapese", "Yapesisk"},
	YID: {"Yiddish", "jiddisch"},
	YOR: {"Yoruba", "joruba"},
	ZAP: {"Zapotec", "Zapotekisk"},
	ZBL: {"Blissymbols", ""},
	ZEN: {"Zenaga", ""},
	ZHA: {"Zhuang", "zhuang"},
	ZUL: {"Zulu", "zulu"},
	ZUN: {"Zuni", ""},
	ZXX: {"no linguistic content", "ingen språklig innhold"},
	ZZA: {"Zazaki", "zazaisk"},
}

var diffCodes = map[Language]string{
		ALB: "sqi",
		ARM: "hye",
		BAQ: "eus",
		BUR: "mya",
		CHI: "zho",
		CZE: "ces",
		DUT: "nld",
		FRE: "fra",
		GEO: "kat",
		GER: "deu",
		GRE: "ell",
		ICE: "isl",
		MAC: "mkd",
		MAO: "mri",
		MAY: "msa",
		NAI: "aqp",
		ORI: "ory",
		PER: "fas",
		RUM: "ron",
		SLO: "slk",
		TIB: "bod",
		WEL: "cym",
	}

var marcToLanguage = map[string]Language{
		"sqi": ALB,
		"hye": ARM,
		"eus": BAQ,
		"mya": BUR,
		"zho": CHI,
		"ces": CZE,
		"nld": DUT,
		"fra": FRE,
		"kat": GEO,
		"deu": GER,
		"ell": GRE,
		"isl": ICE,
		"mkd": MAC,
		"mri": MAO,
		"msa": MAY,
		"aqp": NAI,
		"ory": ORI,
		"fas": PER,
		"ron": RUM,
		"slk": SLO,
		"bod": TIB,
		"cym": WEL,
	}


// ParseLanguageFromMarc parses the given string and returns a Language if
// it matches a known 3-letter Marc language code.
func ParseLanguageFromMarc(s string) (Language, error) {
	if lang, ok := marcToLanguage[s]; ok {
		return lang, nil
	}

	return ParseLanguage(s)
}

// ParseLanguage parses the given string and returns a Language if
// it matches a known 3-letter ISO 639-3 language code.
func ParseLanguage(s string) (Language, error) {
	if _, ok := languageLabels[Language(s)]; ok {
		return Language(s), nil
	}
	return "", fmt.Errorf("iso6393: unknown language code: %q", s)
}

// Code returns the ISO 639-3 language code for the Language.
func (l Language) Code() string {
	return string(l)
}

func (l Language) URI() string {
	return "iso6393/"+string(l)
}

// MarcCode returns the Marc language code for the Language.
// In all but a few cases, the ISO and Marc codes are identical.
func (l Language) MarcCode() string {
	if code, ok := diffCodes[l]; ok {
		return code
	}
	return string(l)
}

// Label returns a string representation of the Marc Language in the desired language.
// Only Norwegian and English are currently supported. If a Norwegian label is required and
// not present, the English label will be returned.
func (l Language) Label(tag language.Tag) string {
	match, _, _ := localizer.Matcher.Match(tag)
	if match == language.Norwegian && languageLabels[l][1] != "" {
		return languageLabels[l][1]
	}
	return languageLabels[l][0]
}

func Options(lang language.Tag) (res [][2]string) {
	match, _, _ := localizer.Matcher.Match(lang)
	i := 0
	if match == language.Norwegian {
		i = 1
	}
	for _, c := range allLanguages {
		label := languageLabels[c][i]
		if label == "" {
			// Fallback to English if missing Norwegian label
			label = languageLabels[c][0]
		}
		res = append(res, [2]string{string(c), label})
	}

	// Sort by label
	sort.Slice(res, func(i, j int) bool {
		return res[i][1] < res[j][1]
	})

	return res
}

	