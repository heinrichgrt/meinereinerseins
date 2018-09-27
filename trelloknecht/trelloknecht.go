package main

import (
	"bytes"
	"context"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adlio/trello"
	"github.com/boombuler/barcode/qr"
	"github.com/coreos/etcd/client"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
	log "github.com/sirupsen/logrus"
)

var (
	configuration = map[string]string{
		// pdf setings
		"fontFamily": "Helvetica",
		"pdfUnitStr": "mm",
		"pdfDocXLen": "100.0",
		"pdfDocYLen": "62.0",
		"pdfMargin":  "3.0",

		"headLineCharsSkip": "82",
		"headLineMaxChars":  "92",
		"printQrCode":       "true",
		"qRCodeSize":        "30.0",
		"qRCodePosX":        "66.0",
		"qRCodePosY":        "25.0",
		"headFontStyle":     "B",
		"headFontSize":      "16.0",
		"headTopMargin":     "5.0",
		"rectX0":            "3.0",
		"rectY0":            "2.0",
		"rectX1":            "95.0",
		"rectY1":            "58.0",

		// trello settings
		"trelloAppKey":       "ad085a9f4dd5cf2d2b558ae45c4ad1f7",
		"trelloToken":        "85e7088cab14a12dee800f262dc15ea6a416157ec2ed1ffe5898234550c9b01b",
		"toPrintedLabelName": "PRINTME_DEVOPS",
		"newLabelAfterPrint": "PRINTED",

		"trelloUserName": "kls_drucker",

		"boardsToWatch":      "DevOps 2020 - Board",
		"printerMedia":       "Custom.62x100mm",
		"printerOrientation": "landscape",
		"printerName":        "Brother_QL_700",
		"tmpDirName":         "",
		"tmpDirPrefix":       "trelloKnecht",
		"numberOfCopiesPrnt": "2",
	}
	//utility vars

	newLabelAfterPrtIDs = make(map[string]string)
	boardNameByID       = make(map[string]string)
	listNameByID        = make(map[string]string)
	labelIDByName       = make(map[string]string)
	cardByFileName      = make(map[string]*trello.Card)
	printedCards        = make([]string, 0)
	// composed vars

	pdfDocDimension = []float64{}
	pdfMargins      = []float64{}
	qRCodePos       = []float64{}
	blackRectPos    = []float64{}
	boardsToWatch   = []string{}
	configFile      = ""

	// printer settings

)

//Resultset  Json for output
type Resultset struct {
	OSCommand            string    `json:"os.cmd"`
	CommandArgs          []string  `json:"cmd.args"`
	Stdout               string    `json:"stdout"`
	Stderr               string    `json:"stderr,omitempty"`
	CmdStarttime         time.Time `json:"cmd.starttime"`
	CMDStoptime          time.Time `json:"cmd.stoptime"`
	DurationSecounds     int       `json:"duration.seconds"`
	SuccessfullExecution bool      `json:"succesful"`
	ErrorStr             string    `json:"errorstr,omitempty"`
}

func getPdfDocDimensionFromString() []float64 {
	r := make([]float64, 0)

	v, _ := strconv.ParseFloat(configuration["pdfDocXLen"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["pdfDocYLen"], 64)
	r = append(r, v)

	return r
}

func getPdfMarginsFromString() []float64 {
	r := make([]float64, 0)
	v, _ := strconv.ParseFloat(configuration["pdfMargin"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["pdfMargin"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["pdfMargin"], 64)
	r = append(r, v)
	return r

}

func getqRCodePosFromString() []float64 {
	r := make([]float64, 0)
	v, _ := strconv.ParseFloat(configuration["qRCodePosX"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["qRCodePosY"], 64)
	r = append(r, v)
	return r

}
func getBlackRectPosFromString() []float64 {
	r := make([]float64, 0)
	v, _ := strconv.ParseFloat(configuration["rectX0"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["rectY0"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["rectX1"], 64)
	r = append(r, v)
	v, _ = strconv.ParseFloat(configuration["rectY1"], 64)

	r = append(r, v)
	return r
}

func checkCommandLineArgs() (bool, string) {

	networked := flag.Bool("networked", false, "get remote config")
	netname := flag.String("netname", "chars", "Metric {chars|words|lines};.")
	debugset := flag.Bool("debug", false, "turn the noise on")
	//configuration["boardsToWatch"] = *flag.String("boards", "DevOps2020 - Board", "board 1, board 2, board n")
	boards := flag.String("boards", "DevOps2020 - Board", "board 1, board 2, board n")
	label := flag.String("label", "", "Label to look for")
	flag.Parse()
	if *debugset {
		log.SetLevel(log.DebugLevel)
	}
	if *label != "" {
		configuration["toPrintedLabelName"] = *label
	}
	if *boards != "" {
		configuration["boardsToWatch"] = *boards
	}
	// TODO the debugger does this wrong
	//*networked = true
	//*netname = "demoprinter"\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\

	return *networked, *netname
}
func fetchIP() string {
	localIPAddr := GetOutboundIP()
	log.Debugf("%v", localIPAddr)
	return localIPAddr.String()

}

/*
func readConfigFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		a := strings.Split(string(scanner.Text()),"=")
		// remove
		fmt.Println("%v", a)

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
*/
func fetchConfiguration() {
	pdfDocDimension = getPdfDocDimensionFromString()
	pdfMargins = getPdfMarginsFromString()
	qRCodePos = getqRCodePosFromString()
	blackRectPos = getBlackRectPosFromString()
	fetchBoardListFromConfig()

}
func putOwnIPtoEtcd(kapi client.KeysAPI, name string, ip string) {
	resp, err := kapi.Set(context.Background(), "/trelloprinter/IPs/"+name, ip, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Debugf("Set is done. Metadata is %q\n", resp)
	}

}
func fetchDefaultConfigFromEtcd(kapi client.KeysAPI) {
	resp, err := kapi.Get(context.Background(), "/trelloprinter/config/default", &client.GetOptions{Recursive: true})
	if err != nil {
		log.Fatal(err)
	} else {

		for _, n := range resp.Node.Nodes {
			log.Debugf("Key: %q, Value: %q\n", n.Key, n.Value)
			configuration[removePathFromKey(n.Key)] = n.Value
		}
	}
}

func fetchOwnConfigurationFromEtcd(kapi client.KeysAPI, printerName string) {
	resp, err := kapi.Get(context.Background(), "/trelloprinter/config/"+printerName, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Fatal(err)
	} else {

		for _, n := range resp.Node.Nodes {
			log.Debugf("Key: %q, Value: %q\n", n.Key, n.Value)
			configuration[removePathFromKey(n.Key)] = n.Value
		}
	}

}
func removePathFromKey(k string) string {
	// there must be a better way...
	//s := regexp.MustCompile("/").Split(k, 5)
	s := strings.Split(k, "/")

	return s[len(s)-1]
}
func fetchBoardListFromConfig() {
	// try this.
	boardsToWatch = strings.Split(configuration["boardsToWatch"], ",")
	log.Debugf("board list: %v", boardsToWatch)
}
func setUpEtcdConnection() client.KeysAPI {
	cfg := client.Config{
		Endpoints:               []string{"http://127.0.0.1:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)

	}
	kapi := client.NewKeysAPI(c)

	return kapi
}
func init() {

	// continue here with flag...
	networked, printername := checkCommandLineArgs()
	fetchConfiguration()

	if networked {
		ip := fetchIP()
		kapi := setUpEtcdConnection()
		fetchDefaultConfigFromEtcd(kapi)
		fetchOwnConfigurationFromEtcd(kapi, printername)
		putOwnIPtoEtcd(kapi, printername, ip)
		fetchBoardListFromConfig()
		log.Error("%v, %v", ip, kapi)

		// write IP to etcd
	}

	dir, err := ioutil.TempDir(os.TempDir(), configuration["tmpDirPrefix"])
	if err != nil {
		log.Fatal(err)
	}
	//defer os.Remove(file.Name())
	log.Debugf(dir)
	configuration["tmpDirName"] = dir
}
func cleanUp(dirName string) {
	os.RemoveAll(dirName)

}
func isPrintedLabelOnBoard(card *trello.Card) bool {
	res := true
	for _, label := range card.Labels {
		if label.Name == configuration["newLabelAfterPrint"] {
			res = false
		}

	}
	return res
}
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
func getPrintedLabelId(board *trello.Board) {
	labels, err := board.GetLabels(trello.Defaults())
	if err != nil {
		log.Fatal("cannot get labels from board: %v", err)
	}
	for _, label := range labels {
		if label.Name == configuration["newLabelAfterPrint"] {
			newLabelAfterPrtIDs[board.ID] = label.ID
		}
	}
}
func swapLabel(cards []*trello.Card) {
	//	x := card.Labels
	//	y := card.IDLabels
	//_, err := card.AddMemberID("testmemberid")
	//if err != nil {
	//	log.Fatalf("add Member: %v\n", err)
	//
	for _, card := range cards {
		r := new(trello.Label)
		//var l card.Labels
		err := card.RemoveIDLabel(labelIDByName[configuration["toPrintedLabelName"]], r)
		if err != nil {
			log.Fatalf("removing  Label : %v with %v \n", configuration["toPrintedLabelName"], err)
		}
		if isPrintedLabelOnBoard(card) {
			err = card.AddIDLabel(newLabelAfterPrtIDs[card.IDBoard])
			if err != nil {
				log.Fatalf("adding Label: %v  with %v\n", configuration["newLabelAfterPrint"], err)
			}
		}
	}
}

func registerQR(pdf *gofpdf.Fpdf, card *trello.Card) {

	key := barcode.RegisterQR(pdf, card.Url, qr.H, qr.Unicode)
	qrSize, _ := strconv.ParseFloat(configuration["qRCodeSize"], 64)
	barcode.BarcodeUnscalable(pdf, key, qRCodePos[0], qRCodePos[1], &qrSize, &qrSize, false)

	// Output:
	// Successfully generated ../../pdf/contrib_barcode_RegisterQR.pdf
}

func shortenStringIfToLong(instring string) string {
	wordList := strings.Split(instring, " ")
	shortendString := ""
	iterator := 0
	headLineLength, err := strconv.Atoi(configuration["headLineCharsSkip"])
	if err != nil {
		log.Fatal("configvalue headLineCharsSkip is nan")
	}
	headLineMaxChars, err := strconv.Atoi(configuration["headLineMaxChars"])
	if err != nil {
		log.Fatal("configvalue headLineMaxChars is nan")
	}
	for len(shortendString) < headLineLength && iterator < len(wordList) {
		if len(shortendString)+len(wordList[iterator]) > headLineMaxChars {
			shortendString += "-bla bla"
			break
		}
		shortendString += " " + wordList[iterator]
		iterator++
	}
	if iterator < len(wordList) {
		shortendString += "..."
	}
	return shortendString
}

func pdfBaseSetup() *gofpdf.Fpdf {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: configuration["pdfUnitStr"],
		Size:    gofpdf.SizeType{Wd: pdfDocDimension[0], Ht: pdfDocDimension[1]},
	})
	pdf.SetMargins(pdfMargins[0], pdfMargins[1], pdfMargins[2])
	pdf.AddPage()
	return pdf
}
func writeLabels(cardList []*trello.Card) []string {
	pdfFileList := make([]string, 0)
	for _, card := range cardList {
		pdf := pdfBaseSetup()
		pdfFileName := writeLabel(pdf, card)
		pdfFileList = append(pdfFileList, pdfFileName)
	}

	return pdfFileList
}

func getBoards(client *trello.Client) []*trello.Board {
	member, err := client.GetMember(configuration["trelloUserName"], trello.Defaults())
	if err != nil {
		log.Fatal("Cannot  get member info from trello")
	}

	boards, err := member.GetBoards(trello.Defaults())
	if err != nil {
		log.Fatal("Cannot get board lists from trello")
	}
	return boards
}
func joinedLabel(card *trello.Card) string {
	labelString := ""
	labelList := make([]string, 0)
	for _, label := range card.Labels {
		if matched, _ := regexp.MatchString("PRINT.*", label.Name); matched == false {
			labelList = append(labelList, label.Name)
		}
	}
	labelString = strings.Join(labelList, ", ")
	return labelString
}

func filterBoards(userBoards []*trello.Board) []*trello.Board {
	boardList := make([]*trello.Board, 0)
	for boardID := range userBoards {
		for watchID := range boardsToWatch {
			if userBoards[boardID].Name == boardsToWatch[watchID] {
				boardList = append(boardList, userBoards[boardID])
			}
		}

	}
	return boardList
}

func writeLabel(pdf *gofpdf.Fpdf, card *trello.Card) string {
	headFontSize, _ := strconv.ParseFloat(configuration["headFontSize"], 64)
	pdf.SetFont(configuration["fontFamily"], configuration["headFontStyle"], headFontSize)

	_, lineHt := pdf.GetFontSize()
	registerQR(pdf, card)
	headTopMargin, _ := strconv.ParseFloat(configuration["headTopMargin"], 64)
	pdf.SetTopMargin(headTopMargin)
	pdf.Rect(blackRectPos[0], blackRectPos[1], blackRectPos[2], blackRectPos[3], "D")
	headerString := card.Name
	htmlString := "<center><bold>" + shortenStringIfToLong(headerString) + "</bold></center>"
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlString)
	htmlString = "<left>" + boardNameByID[card.IDBoard] + " | " + listNameByID[card.IDList] + "</left>"
	pdf.SetFont("Times", "I", 8)
	pdf.SetAutoPageBreak(false, 0.0)
	_, lineHt = pdf.GetFontSize()
	lowerpos := lineHt + 6
	pdf.SetY(-lowerpos)
	html = pdf.HTMLBasicNew()
	html.Write(lineHt, htmlString)
	lowerx := pdf.GetX()
	htmlString = "<right>" + joinedLabel(card) + "</right>"
	pdf.SetX(lowerx + 1)
	pdf.SetY(-lowerpos)
	html = pdf.HTMLBasicNew()
	html.Write(lineHt, htmlString)
	fileName := configuration["tmpDirName"] + "/" + getUUID() + ".pdf"
	cardByFileName[fileName] = card

	err := pdf.OutputFileAndClose(fileName)

	if err != nil {
		log.Error("cannot create pdf-file %v", err)

	}
	return fileName
}
func getUUID() string {
	uuid := uuid.New()
	return uuid.String()

}
func boarListIDsToNames(board *trello.Board) {
	lists, _ := board.GetLists(trello.Defaults())
	for _, list := range lists {

		listNameByID[list.ID] = list.Name

	}

}
func getLabels() []*trello.Card {
	cardList := make([]*trello.Card, 0)
	client := trello.NewClient(configuration["trelloAppKey"], configuration["trelloToken"])
	boards := getBoards(client)

	for _, board := range filterBoards(boards) {
		boarListIDsToNames(board)
		getPrintedLabelId(board)
		boardNameByID[board.ID] = board.Name
		cardList = append(cardList, getMatchingCardsFromBoard(board)...)
	}

	return cardList

}

func getMatchingCardsFromBoard(board *trello.Board) []*trello.Card {
	cardList := make([]*trello.Card, 0)

	cards, err := board.GetCards(trello.Defaults())
	if err != nil {
		log.Fatal("cannot get cards from board")
	}
	for _, card := range cards {

		for _, label := range card.Labels {
			log.Debugf("label %v on %v", label, card)
			if label.Name == configuration["toPrintedLabelName"] {
				cardList = append(cardList, card)
				labelIDByName[label.Name] = label.ID
			}
		}
	}
	return cardList
}

func (r *Resultset) execCommand() {
	cmd := exec.Command(r.OSCommand, r.CommandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	r.SuccessfullExecution = true
	r.CmdStarttime = time.Now()

	err := cmd.Run()
	r.Stdout = string(stdout.Bytes())
	r.Stderr = string(stderr.Bytes())
	r.CMDStoptime = time.Now()
	r.DurationSecounds = int(r.CMDStoptime.Unix() - r.CmdStarttime.Unix())
	if err != nil {
		//todo: log.Fatalf("cmd.Run() failed with %s\n", err)
		log.Errorln("Command failed %v err: ", err)
		r.SuccessfullExecution = false
		r.ErrorStr = err.Error()

	}
	return
}

func printLabels(pdfList []string) {
	for _, pdf := range pdfList {
		commandResult := new(Resultset)
		commandResult.OSCommand = "/usr/bin/lp"
		commandResult.CommandArgs = []string{"-o", "media=" + configuration["printerMedia"], "-o",
			configuration["printerOrientatio"], "-n", configuration["numberOfCopiesPrnt"], "-d", configuration["printerName"], pdf}
		commandResult.execCommand()
		if commandResult.SuccessfullExecution == true {
			printedCards = append(printedCards, pdf)
		}

	}

}

func main() {
	defer cleanUp(configuration["tmpDirName"])
	for {
		cardList := getLabels()
		pdfFileList := writeLabels(cardList)
		printLabels(pdfFileList)
		swapLabel(cardList)
		time.Sleep(60 * time.Second)
	}

}
