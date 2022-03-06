package main

import (
	"DiscordGanggAlarmBot/DiscordBotCore"
	"DiscordGanggAlarmBot/SolSMSCore"
	"DiscordGanggAlarmBot/TwitchCore"
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/mattn/go-tty"
	"log"
	"os"
	"strings"
	"time"
)

func GetDBTable(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	rdr := csv.NewReader(bufio.NewReader(file))
	rows, err := rdr.ReadAll()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return rows, nil
}

func UpdateDBTable(filename string, dataTable [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	wr := csv.NewWriter(bufio.NewWriter(file))

	for _, row := range dataTable {
		wr.Write([]string{row[0], row[1], row[2], row[3]})
		wr.Flush()
	}
	file.Close()
	return nil
}

func GetKeyEvent() rune {
	var ret rune

	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if r != 0 {
			ret = r
			break
		}
	}

	return ret
}

var DB [][]string

func DBDataRemove(slice [][]string, s int) [][]string {
	return append(slice[:s], slice[s+1:]...)
}

func MessageHandler(session *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == session.State.User.ID {
		return
	}

	msgSplit := strings.Split(mc.Content, " ")

	if msgSplit[0] == "!GanggAlarmBot" {
		if msgSplit[0] == "!GanggAlarmBot" && len(msgSplit) == 1 {
			session.ChannelMessageSend(mc.ChannelID, "[강지 방송 알람봇] \n등록하시게 되면 강지 방송이 켜졌을 때 멘션 알람을 드립니다. 필요하시다면 원하시는 전화번호로 SMS도 날려드려요.\n [가이드]\n - !GanggAlarmBot registerMe : 사용자 등록.\n - !GanggAlarmBot unregisterMe : 사용자 등록 해제.\n - !GanggAlarmBot registerSMS 01012341234: 입력된 전화번호로 SMS 서비스 등록.")
			return
		}

		switch msgSplit[1] {
		case "registerMe":
			var bExist bool
			var userId int
			var row1 []string
			for userId, row1 = range DB {
				if row1[0] == mc.Author.ID {
					if strings.Contains(row1[1], mc.ChannelID) {
						session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 이미 등록되어 있습니다.")
					}
					bExist = true
					break
				}
			}
			if !bExist {
				DB = append(DB, []string{mc.Author.ID, mc.ChannelID, "n", "0"})
				err := UpdateDBTable("UserTable.csv", DB)
				if err != nil {
					fmt.Println("DB > Update DB fail. ", err)
				}
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 신규 사용자 등록되었습니다.")
			}
			if bExist && !strings.Contains(row1[1], mc.ChannelID) {
				DB[userId][1] += "-" + mc.ChannelID
				err := UpdateDBTable("UserTable.csv", DB)
				if err != nil {
					fmt.Println("DB > Update DB fail. ", err)
				}
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 신규 사용자 등록되었습니다.")
			}
		case "registerSMS":
			if len(msgSplit) == 3 {
				var bExist bool
				var userId int
				var rowtemp []string
				for userId, rowtemp = range DB {
					if rowtemp[0] == mc.Author.ID {
						bExist = true
						break
					}
				}
				if bExist {
					if DB[userId][2] == "n" {
						DB[userId][2] = "y"
						DB[userId][3] = msgSplit[2]
						_ = UpdateDBTable("UserTable.csv", DB)
						_, _ = session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스가 등록되었습니다.")
					} else if DB[userId][2] == "y" {
						DB[userId][3] = msgSplit[2]
						UpdateDBTable("UserTable.csv", DB)
						session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스 등록 정보를 변경하였습니다.")
					}
				} else {
					session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록 실패. !GanggAlarmBot registerMe가 먼저 선행되어야 합니다.")
				}
			} else {
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록 실패. !GanggAlarmBot registerSMS 01011112222와 같이 전화번호를 추가로 입력해주세요.")
			}
		case "unregisterMe":
			var bExist bool
			var userId int
			var row1 []string
			for userId, row1 = range DB {
				if row1[0] == mc.Author.ID {
					DB = DBDataRemove(DB, userId)
					session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록정보가 삭제되었습니다.")
					bExist = true
					break
				}
			}
			if !bExist {
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록되어 있지 않습니다.")
			}
		}
	}
}

func main() {
	fmt.Println("---< Gangg Alarm Bot >---")

	var solSMSCore *SolSMSCore.SolSMSCore = &SolSMSCore.SolSMSCore{}
	var twitchCore *TwitchCore.TwitchCore = &TwitchCore.TwitchCore{}
	var discordBot *DiscordBotCore.DiscordBotCore = &DiscordBotCore.DiscordBotCore{}

	/* Initialize */
	solSMSCore.Initialize()
	fmt.Println("solSMSCore > Initialize done.")
	err := twitchCore.Initialize("---twitch client key---", "---twitch secret key---")
	if err != nil {
		fmt.Println("twitchCore > Initialize error. ", err)
		return
	}
	fmt.Println("twitchCore > Initialize done.")
	err = discordBot.Initialize("---discord bot token---", MessageHandler)
	fmt.Println("discordBot > Initialize done.")
	DB, err = GetDBTable("UserTable.csv")
	if err != nil {
		fmt.Println("DB > Data read fail. ", err)
		return
	}
	fmt.Println("DB > Connected.")

	/* Run */
	discordBot.Activate()
	fmt.Println("> Bot running... Press Q to exit.")
	var alarmSwitch map[string]bool = make(map[string]bool)
	for {
		DB, _ = GetDBTable("UserTable.csv")
		for _, row := range DB {
			_, exist := alarmSwitch[row[0]]
			if !exist {
				alarmSwitch[row[0]] = false
			}
		}
		for key, _ := range alarmSwitch {
			var bGhost bool = true
			for _, row := range DB {
				if row[0] == key {
					bGhost = false
					break
				}
			}
			if bGhost {
				delete(alarmSwitch, key)
			}
		}
		if twitchCore.IsStreamerLive("rkdwl12") {
			fmt.Println("twitchCore > rkdwl12 stream on.")
			for index, _ := range DB {
				if alarmSwitch[DB[index][0]] == false {
					fmt.Println("Core > Stream Alarm for " + DB[index][0])
					// Send discord chat.
					discordChannelList := strings.Split(DB[index][1], "-")
					for _, channel := range discordChannelList {
						discordBot.SendChannelMessage(channel, "<@"+DB[index][0]+"> 감자의 생방송이 시작대떠 :P")
					}
					// Send SMS.
					if DB[index][2] == "y" {
						solSMSCore.SendSMS(DB[index][3], "---phone number---", "감자의 생방송이 시작대떠 :P")
					}
					alarmSwitch[DB[index][0]] = true
				}
			}
		} else {
			fmt.Println("twitchCore > rkdwl12 stream off.")
			for key, value := range alarmSwitch {
				if value {
					alarmSwitch[key] = false
				}
			}
		}
		// == time.Sleep(20 * time.Second)
		var deltaTime time.Duration
		var startTime time.Time
		for deltaTime.Seconds() < 20 {
			startTime = time.Now()
			input := GetKeyEvent()
			if input == 113 {
				/* Destroy */
				discordBot.Destroy()
				fmt.Println("discordBot > Destroy")
				return
			}
			deltaTime -= startTime.Sub(time.Now())
		}
	}

	/* Destroy */
	discordBot.Destroy()
	fmt.Println("discordBot > Destroy")
}
