package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/adlio/trello"
	"github.com/boombuler/barcode/qr"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/barcode"
	log "github.com/sirupsen/logrus"
)

var (
	// pdf setings
	fontFamily        = "Helvetica"
	pdfUnitStr        = "mm"
	pdfDocDimension   = []float64{100.0, 62.0}
	pdfMargins        = []float64{3.0, 3.0, 3.0}
	headLineCharsSkip = 82
	headLineMaxChars  = 92
	qRCodeSize        = 30.0
	qRCodePos         = []float64{68.0, 25.0}
	headFontStyle     = "B"
	headFontSize      = 14.0
	headTopMargin     = 5.0
	blackRectPos      = []float64{2.0, 2.0, 96.0, 58.0}
	// trello settings
	trelloAppKey       = "ad085a9f4dd5cf2d2b558ae45c4ad1f7"
	trelloToken        = "85e7088cab14a12dee800f262dc15ea6a416157ec2ed1ffe5898234550c9b01b"
	toPrintedLabelName = "PRINTME_NDL"
	trelloUserName     = "kls_drucker"
	boardsToWatch      = []string{"DevOps 2020 - Board", "SAP Backlog KLS"}
	labelName          = "PRINTME_NDL"

	//utility vars
	boardNameByID = make(map[string]string)
	listNameByID  = make(map[string]string)

	// system settings
	tmpDirPrefix = "trelloKnecht"
	// printer settings
	printerMedia       = "Custom.62x100mm"
	printerOrientation = "landscape"
	printerName        = "Brother_QL_700"
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

func startUp() {
	file, err := ioutil.TempFile("dir", tmpDirPrefix)
	if err != nil {
		log.Fatal(err)
	}
	//defer os.Remove(file.Name())

	fmt.Println(file.Name())

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
func getMarkedCardsByBoard(board *trello.Board) []*trello.Card {
	var matchingCards []*trello.Card
	cards, err := board.GetCards(trello.Defaults())
	if err != nil {
		log.Error("cannot get Cards from Board: %v", board.Name)
	}
	for cardID := range cards {
		labels := cards[cardID].Labels
		for labelID := range labels {
			log.Debug("label: %v", labels[labelID].Name)
			fmt.Printf("log %v\n", labels[labelID].Name)
			if labels[labelID].Name == toPrintedLabelName {

				matchingCards = append(matchingCards, cards[cardID])
			}
		}
	}
	return matchingCards
}
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
		extendedPdf := writeLabel(pdf, card)
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
		x := userBoards[boardID].Name
		for watchID := range boardsToWatch {
			if x == boardsToWatch[watchID] {
				boardList = append(boardList, userBoards[boardID])
			}
		}

	}
	return boardList
}
func writeLabel(pdf *gofpdf.Fpdf, card *trello.Card) string {

	pdf.SetFont(fontFamily, headFontStyle, headFontSize)
	extended := make([]*trello.Card, 0)
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
	err := pdf.OutputFileAndClose("/Users/heinrich/card.pdf")

	if err != nil {
		log.Error("cannot create pdf-file %v", err)

	}
	return "defineThisLater"
	// add code for the card.
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

	list, _ := client.GetList("5a53520750c9f99b20f6c7b9", trello.Defaults())
	fmt.Printf("%v", list)
	// filteredBoards := filterBoards(boards)
	for _, board := range filterBoards(boards) {
		boarListIDsToNames(board)
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
			if label.Name == labelName {
				cardList = append(cardList, card)
			}
		}
	}
	return cardList
}

/* for cardID := range cards {
					//	fmt.Printf("card %v", cards[card])
					for labelId := range cards[cardID].IDLabels {
						// fmt.Printf("label: %v\n", cards[card].IDLabels[labelId])
						x, _ := client.GetLabel(cards[cardID].IDLabels[labelId], trello.Defaults())
						cardno++
						fmt.Printf("card no: %v, label: %v\n", cardno, x.Name)

						//fmt.Printf("label %v\n", labelId)
					}

				}

			}
		card.Labels

	}

}
*/
func execCommand(CommandOutPut *Resultset) {
	//func execCommand(extcmd string, args []string) error, string, string{

	cmd := exec.Command(CommandOutPut.OSCommand, CommandOutPut.CommandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	CommandOutPut.SuccessfullExecution = true
	CommandOutPut.CmdStarttime = time.Now()

	err := cmd.Run()
	CommandOutPut.Stdout = string(stdout.Bytes())
	CommandOutPut.Stderr = string(stderr.Bytes())
	CommandOutPut.CMDStoptime = time.Now()
	CommandOutPut.DurationSecounds = int(CommandOutPut.CMDStoptime.Unix() - CommandOutPut.CmdStarttime.Unix())
	if err != nil {
		//log.Fatalf("cmd.Run() failed with %s\n", err)
		log.Errorln("Command failed %v err: ", err)
		CommandOutPut.SuccessfullExecution = false
		CommandOutPut.ErrorStr = err.Error()

	}

	//	fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

	return
}

func printLabels(pdfList []string) {
	for _, pdf := range pdfList {
		commandResult := new(Resultset)
		commandResult.OSCommand = "/usr/bin/lp"
		commandResult.CommandArgs = []string{"-o", "media=" + printerMedia, "-o", printerOrientation, "-d", printerName, pdf}

	}
}
func main() {

	cardList := getLabels()
	pdfFileList := writeLabels(cardList)
	printLabels(pdfFileList)
}
