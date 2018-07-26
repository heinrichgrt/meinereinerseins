package main

import (
	"bytes"
	"fmt"
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
	// pdf setings
	fontFamily = "Helvetica"
	pdfUnitStr = "mm"
	pdfDocXLen = 100.0
	pdfDocYLen = 62.0
	pdfMargin  = 3.0

	headLineCharsSkip = 82
	headLineMaxChars  = 92
	qRCodeSize        = 30.0
	qRCodePosX        = 68.0
	qRCodePosY        = 25.0
	headFontStyle     = "B"
	headFontSize      = 14.0
	headTopMargin     = 5.0
	rectX0            = 3.0
	rectY0            = 2.0
	rectX1            = 95.0
	rectY1            = 58.0

	// trello settings
	trelloAppKey       = "ad085a9f4dd5cf2d2b558ae45c4ad1f7"
	trelloToken        = "85e7088cab14a12dee800f262dc15ea6a416157ec2ed1ffe5898234550c9b01b"
	toPrintedLabelName = "PRINTME_DEVOPS"
	newLabelAfterPrint = "PRINTED"

	trelloUserName = "kls_drucker"

	boardsToWatch = []string{"DevOps 2020 Themen und Ideen"}

	//utility vars
	newLabelAfterPrtIDs = make(map[string]string)
	boardNameByID       = make(map[string]string)
	listNameByID        = make(map[string]string)
	labelIDByName       = make(map[string]string)
	cardByFileName      = make(map[string]*trello.Card)
	printedCards        = make([]string, 0)
	// composed vars
	pdfDocDimension = []float64{pdfDocXLen, pdfDocYLen}
	pdfMargins      = []float64{pdfMargin, pdfMargin, pdfMargin}
	qRCodePos       = []float64{qRCodePosX, qRCodePosY}
	blackRectPos    = []float64{rectX0, rectY0, rectX1, rectY1}
	// system settings
	tmpDirPrefix = "trelloKnecht"
	// printer settings
	printerMedia       = "Custom.62x100mm"
	printerOrientation = "landscape"
	printerName        = "Brother_QL_700"
	tmpDirName         = ""
	numberOfCopiesPrnt = 2
)

//Resultset  Json for output
type Resultset struct {
	OSCommand            string    `json:"os.cmd"`
	CommandArgs          []string  `json:"cmd.args"`
	Stdout               string    `json:"stdout"`
	Stderr               string    `json:"stderr,omitempty"`
	CmdStarttime         time.Time `json:"cmd.starttime"`
	CMDStoptime          time.Time `json:"=md.stoptime"`
	DurationSecounds     int       `json:"duration.seconds"`
	SuccessfullExecution bool      `json:"succesful"`
	ErrorStr             string    `json:"errorstr,omitempty"`
}
type empt struct {
	name string `json:"ignore"`
}

// API extensions:

/*
func deleteLabel(card *trello.Card, labelID string `json:"ignore"`) error {
	path := fmt.Sprintf("cards/%s", card.ID)
	return card.client.Delete(path, Arguments{"pos": fmt.Sprintf("%f", newPos)}, c)
	card.AddIdLabel("PRINTED")
}
*/
func checkCommandLineArgs() bool {
	argsWithProg := os.Args
	networked := false
	if len(argsWithProg) > 1 && argsWithProg[1] == "-n" {
		networked = true
	}
	return networked
}
func fetchIP() string {
	localIPAddr := GetOutboundIP()
	fmt.Println("%v", localIPAddr)
	return localIPAddr.String()

}
func fetchDefaultConfigFromEtcd(kapi client.KeysAPI) {
	kapi.Set(weitermachen)

}
func setUpEtcdConnection() client.KeysAPI {
	cfg := client.Config{
		Endpoints: []string{"http://127.0.0.1:2379"},
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)

	}
	kapi := client.NewKeysAPI(c)
	/*
		 set "/foo" key with "bar" value
		log.Print("Setting '/foo' key with 'bar' value")
		resp, err := kapi.Set(context.Background(), "/foo", "bar", nil)
		//client.New(cfg)
		if err != nil {
			log.Fatal(err)
		} else {
			// print common key info
			log.Printf("Set is done. Metadata is %q\n", resp)
		}
	*/
	return kapi
}
func init() {
	networked := checkCommandLineArgs()
	if networked {
		ip := fetchIP()
		kapi := setUpEtcdConnection()
		log.Error("%v, %v", ip, kapi)

	}

	dir, err := ioutil.TempDir(os.TempDir(), tmpDirPrefix)
	if err != nil {
		log.Fatal(err)
	}
	//defer os.Remove(file.Name())
	fmt.Println(dir)
	tmpDirName = dir
}
func cleanUp(dirName string) {
	os.RemoveAll(dirName)

}
func isPrintedLabelOnBoard(card *trello.Card) bool {
	res := true
	for _, label := range card.Labels {
		if label.Name == newLabelAfterPrint {
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
		fmt.Printf("%v", label)
		if label.Name == newLabelAfterPrint {
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
		err := card.RemoveIDLabel(labelIDByName[toPrintedLabelName], r)
		if err != nil {
			log.Fatalf("removing  Label : %v with %v \n", toPrintedLabelName, err)
		}
		if isPrintedLabelOnBoard(card) {
			err = card.AddIDLabel(newLabelAfterPrtIDs[card.IDBoard])
			if err != nil {
				log.Fatalf("adding Label: %v  with %v\n", newLabelAfterPrint, err)
			}
		}
	}
	/* newLabelList := make([]string, 0)

	for _, labelId := range y {

		if labelId != labelIDByName[toPrintedLabelName] {

			//	matchingCards = append(matchingCards, cards[cardID])

			newLabelList = append(newLabelList, labelId)
		}
	}
	//
	t := make([]*trello.Label, 0)
	card.Labels = t
	//
	card.IDLabels = newLabelList
	error := card.Update(trello.Defaults())

	log.Info("%v", error)
	log.Info("%v, %v", x, y)
	*/
}
func registerQR(pdf *gofpdf.Fpdf, card *trello.Card) {

	key := barcode.RegisterQR(pdf, card.Url, qr.H, qr.Unicode)

	barcode.BarcodeUnscalable(pdf, key, qRCodePos[0], qRCodePos[1], &qRCodeSize, &qRCodeSize, false)

	// Output:
	// Successfully generated ../../pdf/contrib_barcode_RegisterQR.pdf
}

func shortenStringIfToLong(instring string) string {
	wordList := strings.Split(instring, " ")
	shortendString := ""
	iterator := 0
	for len(shortendString) < headLineCharsSkip && iterator < len(wordList) {
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

/*
func getMarkedCardsByBoard(board *trello.Board) []*trello.Card {
	var matchingCards []*trello.Card
	cards, err := board.GetCards(trello.Defaults())
	if err != nil {
		log.Error("cannot get Cards from Board: %v", board.Name)
	}
	for cardID := range cards {
		labels := cards[cardID].Labels
		for labelID := range labels {
			labelIDByName[labels[labelID].Name] = labels[labelID].ID
			if labels[labelID].Name == toPrintedLabelName {
				matchingCards = append(matchingCards, cards[cardID])
			}
		}
	}
	return matchingCards
}
*/
func pdfBaseSetup() *gofpdf.Fpdf {
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: pdfUnitStr,
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
	member, err := client.GetMember(trelloUserName, trello.Defaults())
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
		fmt.Printf("id: %v, board: %v", boardID, userBoards[boardID])
		for watchID := range boardsToWatch {
			if userBoards[boardID].Name == boardsToWatch[watchID] {
				boardList = append(boardList, userBoards[boardID])
			}
		}

	}
	return boardList
}
func writeLabel(pdf *gofpdf.Fpdf, card *trello.Card) string {

	pdf.SetFont(fontFamily, headFontStyle, headFontSize)

	_, lineHt := pdf.GetFontSize()
	registerQR(pdf, card)
	pdf.SetTopMargin(headTopMargin)
	pdf.Rect(blackRectPos[0], blackRectPos[1], blackRectPos[2], blackRectPos[3], "D")
	headerString := card.Name
	htmlString := "<center>" + shortenStringIfToLong(headerString) + "</center>"
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
	fileName := tmpDirName + "/" + getUUID() + ".pdf"
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
	client := trello.NewClient(trelloAppKey, trelloToken)
	boards := getBoards(client)

	for _, board := range filterBoards(boards) {
		boarListIDsToNames(board)
		getPrintedLabelId(board)
		boardNameByID[board.ID] = board.Name
		cardList = append(cardList, getMatchingCardsFromBoard(board)...)
	}

	fmt.Printf("boards: %v", boards)
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
			fmt.Println("label %v on %v", label, card)
			if label.Name == toPrintedLabelName {
				cardList = append(cardList, card)
				labelIDByName[label.Name] = label.ID
			}
		}
	}
	return cardList
}

func (r *Resultset) execCommand() {
	//func execCommand(extcmd string, args []string) error, string, string{

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
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		log.Errorln("Command failed %v err: ", err)
		r.SuccessfullExecution = false
		r.ErrorStr = err.Error()

	}

	//	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	return
}

func printLabels(pdfList []string) {
	for _, pdf := range pdfList {
		commandResult := new(Resultset)
		commandResult.OSCommand = "/usr/bin/lp"
		commandResult.CommandArgs = []string{"-o", "media=" + printerMedia, "-o",
			printerOrientation, "-n", strconv.Itoa(numberOfCopiesPrnt), "-d", printerName, pdf}
		commandResult.execCommand()
		fmt.Printf("%v", commandResult)
		if commandResult.SuccessfullExecution == true {
			printedCards = append(printedCards, pdf)
		}

	}

}

func main() {
	defer cleanUp(tmpDirName)
	cardList := getLabels()
	pdfFileList := writeLabels(cardList)
	printLabels(pdfFileList)
	swapLabel(cardList)

}
