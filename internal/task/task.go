package task

import (
	"fmt"
	"github.com/kelvinzer0/imail/internal/conf"
	"github.com/kelvinzer0/imail/internal/db"
	"github.com/kelvinzer0/imail/internal/log"
	"github.com/kelvinzer0/imail/internal/smtpd"
	"github.com/kelvinzer0/imail/internal/tools/cron"
	"github.com/kelvinzer0/imail/internal/tools/mail"
)

var c = cron.New()

func TaskQueueeSendMail() {

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		return
	}

	postmaster := fmt.Sprintf("postmaster@%s", from)
	result, err := db.MailSendListForStatus(2, 1)
	if err != nil {
		log.Errorf("MailSendListForStatus error: %v", err)
		return
	}
	if len(result) == 0 {

		result, err = db.MailSendListForStatus(0, 1)
		if err != nil {
			log.Errorf("MailSendListForStatus error: %v", err)
			return
		}
		for _, val := range result {
			db.MailSetStatusById(val.Id, 2)

			content, err := db.MailContentRead(val.Uid, val.Id)
			if err != nil {
				continue
			}
			err = smtpd.Delivery("", val.MailFrom, val.MailTo, []byte(content))

			if err != nil {
				content, _ := mail.GetMailReturnToSender(postmaster, val.MailFrom, val.MailTo, content, err.Error())
				db.MailPush(val.Uid, 1, postmaster, val.MailFrom, content, 1, false)
			}
			db.MailSetStatusById(val.Id, 1)
		}
	}
}

func TaskRspamdCheck() {

	result, err := db.MailListForRspamd(1)
	if err != nil {
		log.Errorf("MailListForRspamd error: %v", err)
		return
	}
	if conf.Rspamd.Enable {
		for _, val := range result {
			content, err := db.MailContentRead(val.Uid, val.Id)
			if err != nil {
				continue
			}
			_, err, score := mail.RspamdCheck(content)
			// fmt.Println("RspamdCheck:", val.Id, err)
			if err == nil {
				db.MailSetIsCheckById(val.Id, 1)
				log.Infof("mail[%d] check is pass! score:%f", val.Id, score)
			} else {
				db.MailSetIsCheckById(val.Id, 1)
				db.MailSetJunkById(val.Id, 1)
				log.Errorf("mail[%d] check is spam! score:%f", val.Id, score)
			}
		}
	}
}

func Init() {

	c.AddFunc("mail send task", "@every 5s", func() { TaskQueueeSendMail() })
	c.AddFunc("mail rspamd check", "@every 10m", func() { TaskRspamdCheck() })

	c.Start()
}

// ListTasks returns all running cron tasks.
func ListTasks() []*cron.Entry {
	return c.Entries()
}
