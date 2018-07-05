package main

import (
	"fmt"
	"regexp"
	"strings"

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
)

func registerQR(pdf *gofpdf.Fpdf) {

	key := barcode.RegisterQR(pdf, "https://trello.com/c/JJRO2z8y/260-die-containerlinux-updates-finden-zu-einer-geeigneten-zeit-statt", qr.H, qr.Unicode)

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
func writeLabels(cardList []*trello.Card) []*gofpdf.Fpdf {
	pdfList := make([]*gofpdf.Fpdf, 0)
	for _, card := range cardList {
		pdf := pdfBaseSetup()
		writeLabel(pdf, card)

	}
	return pdfList
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
func writeLabel(pdf *gofpdf.Fpdf, card *trello.Card) {

	pdf.SetFont(fontFamily, headFontStyle, headFontSize)
	_, lineHt := pdf.GetFontSize()
	registerQR(pdf)
	pdf.SetTopMargin(headTopMargin)
	xx, yy := pdf.GetXY()

	pdf.Rect(blackRectPos[0], blackRectPos[1], blackRectPos[2], blackRectPos[3], "D")
	headerString := card.Name

	//columnName := card.Labels
	columnName := card.IDList
	htmlString := "<center>" + shortenStringIfToLong(headerString) + "</center>"

	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlString)

	htmlString = "<left>" + boardNameByID[card.IDBoard] + " | " + columnName + "</left>"
	xx, yy = pdf.GetXY()
	//fmt.Printf("x %v und y %v", xx, yy)
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
	fmt.Printf("x %v und y %v ", xx, yy)

	/* pdf.SetFooterFunc(func {
		pdf.SetY(-15)
		pdf.SetFont(fontFamily, "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Unten auf der Seite"),
			"", 0, "C", false, 0, "")
	})
	*/

	err := pdf.OutputFileAndClose("/Users/heinrich/card.pdf")
	if err != nil {
		log.Error("cannot create pdf-file %v", err)

	}

}
func boarListIDsToNames(board *trello.Board) {
	list, _ := board.GetLists(trello.Defaults())
	fmt.Println("%v", list)
}
func getLabels() []*trello.Card {
	cardList := make([]*trello.Card, 0)
	client := trello.NewClient(trelloAppKey, trelloToken)
	boards := getBoards(client)
	list, _ := client.GetList("5a53520750c9f99b20f6c7b9", trello.Defaults())
	fmt.Printf("%v", list)
	// filteredBoards := filterBoards(boards)
	for _, board := range filterBoards(boards) {
		boardNameByID[board.ID] = board.Name
		cardList = append(cardList, getMatchingCardsFromBoard(board)...)
	}

	fmt.Printf("boards: %v", boards)
	return cardList

}
func checkCardForLabel(cardList []*trello.Card) {

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

func main() {

	cardList := getLabels()
	writeLabels(cardList)
	// todo: writeLabels()

	/*

		board, err := client.GetBoard("5a4cafbaac838c7713a3a7e3", trello.Defaults())
		if err != nil {
			log.Error("cannot get Board: %v", "change")
		}
		cardsToPrint := getMarkedCardsByBoard(board)

		for cardID := range cardsToPrint {
			fmt.Printf("card name %v", cardsToPrint[cardID])
	*/

	//}

	//
	/*

		// new FPDF('P','mm',array(100,150));
		//label, err := client.GetLabel("4eea4ff", Defaults())

		if err != nil {
			log.Errorf("cann not get board with id:  ")
		}
		/*
			cards, err := board.GetCards(trello.Defaults())

			cardno := 0
			/* for _, cards := range cards {

				// GetCards makes an API call to /lists/:id/cards using credentials from `client`
				//cards, err := board.GetCards(trello.Defaults())
				if err != nil {
					// Handle error
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
			fmt.Printf("baord: %v\n", board)
	*/
}
