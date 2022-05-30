package main

import (
	"DiscordGanggAlarmBot/DiscordBotCore"
	"DiscordGanggAlarmBot/SolSMSCore"
	"DiscordGanggAlarmBot/TwitchCore"
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
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

var DB [][]string

type prCell struct {
	userId    string
	userPhone string
}

var prDB map[string]prCell = make(map[string]prCell)

func DBDataRemove(slice [][]string, s int) [][]string {
	return append(slice[:s], slice[s+1:]...)
}

func MessageHandler(session *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == session.State.User.ID {
		return
	}

	msgSplit := strings.Split(mc.Content, " ")

	if msgSplit[0] == "!gbot" {
		if msgSplit[0] == "!gbot" && len(msgSplit) == 1 {
			session.ChannelMessageSend(mc.ChannelID, "[ 강지 방송 알람봇 GBOT ] \n강지 방송을 놓치는 일이 없도록 방송 알람을 제공하는 gbot 입니다. 등록하시게 되면 기본적으로 디스코드 멘션 메시지로 알람을 드립니다. SMS를 추가로 등록하신다면 휴대폰 메시지로도 알람을 받아보실 수 있습니다.\n\n*팬심으로 만들어진 비영리 프로그램으로 SMS 수신에 있어 비용이 드는 관계로 최대 20분까지만 등록하실 수 있습니다.*\n\n [메뉴얼]\n > !gbot -n : 사용자 등록.\n > !gbot -d : 사용자 등록 해제.\n > !gbot -p 01012341234: 입력된 전화번호로 SMS 서비스 등록.")
			return
		}

		switch msgSplit[1] {
		case "-n":
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
		case "-p":
			if len(msgSplit) == 3 {
				var smsCount int = 0
				for _, data := range DB {
					if data[2] == "y" {
						smsCount++
					}
				}
				if smsCount >= 20 {
					session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스 신청이 마감되었습니다ㅜㅠ. 개발자가 돈이 많아지면 더 늘릴 수 있도록 하겠습니다.")
					return
				}

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
						//DB[userId][2] = "y"
						//DB[userId][3] = msgSplit[2]
						//_ = UpdateDBTable("UserTable.csv", DB)
						//_, _ = session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스가 등록되었습니다.")
						random := rand.New(rand.NewSource(time.Now().UnixNano()))
						randomNumber := strconv.Itoa(random.Intn(100000)) + strconv.Itoa(len(prDB))
						prDB[randomNumber] = prCell{DB[userId][0], msgSplit[2]}
						solSMSCore.SendSMS(msgSplit[2], "01011112222", "GBOT SMS 서비스 등록 인증번호입니다. \n"+randomNumber)
						session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 신청하신 번호로 인증번호를 발송하였습니다.\n!gbot -pp (인증번호) 를 입력하여 인증을 마무리해주세요.")
					} else if DB[userId][2] == "y" {
						//DB[userId][3] = msgSplit[2]
						//UpdateDBTable("UserTable.csv", DB)
						//session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스 등록 정보를 변경하였습니다.")
						random := rand.New(rand.NewSource(time.Now().UnixNano()))
						randomNumber := strconv.Itoa(random.Intn(100000)) + strconv.Itoa(len(prDB))
						prDB[randomNumber] = prCell{DB[userId][0], msgSplit[2]}
						solSMSCore.SendSMS(msgSplit[2], "01011112222", "GBOT SMS 서비스 등록 인증번호입니다. \n"+randomNumber)
						session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 신청하신 번호로 인증번호를 발송하였습니다.\n!gbot -pp (인증번호) 를 입력하여 인증을 마무리해주세요.")
					}
				} else {
					session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록 실패. !gbot -n 이 먼저 선행되어야 합니다.")
				}
			} else {
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 등록 실패. !gbot -p 01011112222와 같이 전화번호를 추가로 입력해주세요.")
			}
		case "-pp":
			if len(msgSplit) == 3 {
				if val, ok := prDB[msgSplit[2]]; ok {
					for idx, data := range DB {
						if data[0] == val.userId {
							DB[idx][2] = "y"
							DB[idx][3] = val.userPhone
							UpdateDBTable("UserTable.csv", DB)
							session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> SMS 서비스가 등록되었습니다.")
							delete(prDB, msgSplit[2])
							break
						}
					}
				} else {
					session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 인증 실패. 유효하지 않은 인증번호입니다.")
				}
			} else {
				session.ChannelMessageSend(mc.ChannelID, "<@"+mc.Author.ID+"> 인증 실패. !gbot -pp (인증번호) 와 같은 형식으로 입력해주세요.")
			}
		case "-d":
			var bExist bool
			var userId int
			var row1 []string
			for userId, row1 = range DB {
				if row1[0] == mc.Author.ID {
					DB = DBDataRemove(DB, userId)
					err := UpdateDBTable("UserTable.csv", DB)
					if err != nil {
						fmt.Println("DB > Update DB fail. ", err)
					}
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

var solSMSCore *SolSMSCore.SolSMSCore = &SolSMSCore.SolSMSCore{}
var twitchCore *TwitchCore.TwitchCore = &TwitchCore.TwitchCore{}
var discordBot *DiscordBotCore.DiscordBotCore = &DiscordBotCore.DiscordBotCore{}

func main() {
	fmt.Println("---< Gangg Alarm Bot >---")

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
	ch := make(chan string)
	go func(ch chan string) {
		// disable input buffering
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		// do not display entered characters on the screen
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		var b = make([]byte, 1)
		for {
			os.Stdin.Read(b)
			ch <- string(b)
		}
	}(ch)

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
						solSMSCore.SendSMS(DB[index][3], "01011112222", "감자의 생방송이 시작대떠 :P")
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
			select {
			case input, _ := <-ch:
				if input == "q" {
					/* Destroy */
					discordBot.Destroy()
					fmt.Println("discordBot > Destroy")
					return
				}
			default:
			}
			time.Sleep(time.Millisecond * 10)
			deltaTime -= startTime.Sub(time.Now())
		}
	}

	/* Destroy */
	discordBot.Destroy()
	fmt.Println("discordBot > Destroy")
}
