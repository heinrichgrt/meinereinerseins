# Trelloknecht
A tool to print Trello cards on a label printer.

There was a python thingie around basically doing the same but this is no longer maintained. So I wrote something replacing it. 

## Introducion
Mark the Trello Card you want to be printed with a label e.g. "PRINTME". This software scans a list of boards, finds the cards with the label, prints them and replaces the label with a new label e.g. "PRINTED". 

## Requirements
- A Trello Board
- A label Printer, we are using Brother QL XXX 
- A go runtime environment 
- A "computer" running this software. 


## Installation 

go get github.com/heinrichgrt/trelloknecht


## Setup
- Add a printing user to your organisation or use an existing technical user.
- Create Access token for this user if not already in place. 
- Invite this user to all the boards you want to print from.
- Choose card label for the state: "To be printed" and "Printed" and create them on every board. E.g. "PRINTME" and "PRINTED". 
- Create a file with the access token. Details below. 
- Edit config.cfg to your needs. 
- Start the software. 
- Add the "PRINTME" to a card. 
- Wait until your label printer prints the label. 
- done. 

