package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/ledongthuc/pdf"
	log "github.com/sirupsen/logrus"
)

// Input - a docx file
// Output - a map of which newspaper article mentions which legal article (and their respective laws)

const logLevel = log.DebugLevel

// A map of legal documents and their correct spelling
var legalDocs = map[string]string{
	"ЗИНЗС":    "ЗИНЗС",    //ЗАКОН ЗА ИЗПЪЛНЕНИЕ НА НАКАЗАНИЯТА И ЗАДЪРЖАНЕТО ПОД СТРАЖА
	"ППЗИНСЗС": "ППЗИНСЗС", //ПРАВИЛНИК ЗА ПРИЛАГАНЕ НА ЗИНС
	"ПИЗНС":    "ЗИНЗС",    // (typo) ЗИНЗС
	"НК":       "НК",       //НАКАЗАТЕЛЕН КОДЕКС
	"НПК":      "НПК",      //НАКАЗАТЕЛНО ПРОЦЕСУАЛЕН КОДЕКС
	"ИПИП":     "ИПИП",     // (термин) Индивидуален План за Изпълнение на Присъда
	"ИСДВР":    "ИСДВР",    // (термин) Инспектор Социална Дейност и Възпитателна Работа (социалните работници)
	"ГДИН":     "ГДИН",     // (термин) Главна Дирекция Изпълнение на Наказанията
	"ППСОРРВ":  "Правила за Прилагане на Системата за Оценка на Риска от Рецидив и Вреди",
	"НОС":      "НОС",  // (термин) Надзорно-охранителен състав
	"ЕСПЧ":     "ЕСПЧ", // (термин) Европейски съд по правата на човека
	"БХК":      "БХК",  // (термин) Български Хелзински Комитет
	"УПО":      "УПО",  // (термин) условно предсрочно освобождаване
	"АПК":      "АПК",  // Административнопроцесуален кодекс
	"КРБ":      "КРБ",  // Конституция на Република България
	"Конституционно Решение": "Конституционно Решение", // (special) конкретни решения които трябва да се внимава и вадят ръчно (включват номер + година)
	"ВАС": "ВАС", // (термин) висш административен съд

}

var issueDates = map[int]time.Time{
	1:  formatIssueDate(2020, 8),
	2:  formatIssueDate(2020, 9),
	3:  formatIssueDate(2020, 10),
	4:  formatIssueDate(2020, 11),
	5:  formatIssueDate(2020, 12),
	6:  formatIssueDate(2021, 1),
	7:  formatIssueDate(2021, 2),
	8:  formatIssueDate(2021, 3),
	9:  formatIssueDate(2021, 4),
	10: formatIssueDate(2021, 5),
	11: formatIssueDate(2021, 6),
	12: formatIssueDate(2021, 7),
	13: formatIssueDate(2021, 8),
	14: formatIssueDate(2021, 9),
	15: formatIssueDate(2021, 10),
	16: formatIssueDate(2021, 11),
	17: formatIssueDate(2021, 12),
	18: formatIssueDate(2022, 1),
	19: formatIssueDate(2022, 2),
	20: formatIssueDate(2022, 3),
	21: formatIssueDate(2022, 4),
	22: formatIssueDate(2022, 5),
}

func main() {

	var issNum int
	flag.IntVar(&issNum, "issue", 0, "issue number between 1 - 22")
	issDate := issueDates[issNum]
	flag.Parse()

	log.SetLevel(logLevel)
	pdf.DebugOn = true
	log.Info("Extracting issue", &issNum)
	issueString, err := loadPdf(fmt.Sprintf("is%d.pdf", issNum))
	if err != nil {
		panic(err)
	}
	issue := extractIssue(issueString, issNum, issDate)
	log.Infof("Finished Extracting issue %d", issue.IssueNum)

	// create wp-api client
	client := WPClient{
		baseUrl:  "http://localhost:10008/wp-json/wp/v2", // example: `http://192.168.99.100:32777/wp-json/wp/v2`
		username: "api-user",
		password: "zAaY NhiE UmIY cnL1 syVC MVSr",
	}

	for header, article := range issue.Articles {
		// TODO CREATE and PUBLISH post with extracted article info
		createdPost := client.CreatePost(issue, header, article)
		log.Debug(createdPost)
	}
}
