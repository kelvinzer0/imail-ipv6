package mail

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/kelvinzer0/imail-ipv6/internal/app/context"
	"github.com/kelvinzer0/imail-ipv6/internal/app/form"
	"github.com/kelvinzer0/imail-ipv6/internal/conf"
	"github.com/kelvinzer0/imail-ipv6/internal/db"
	"github.com/kelvinzer0/imail-ipv6/internal/tools"
	tmail "github.com/kelvinzer0/imail-ipv6/internal/tools/mail"
	"github.com/kelvinzer0/imail-ipv6/internal/tools/paginater"
	"github.com/midoks/mcopa"
)

const (
	MAIL             = "mail/list"
	MAIL_NEW         = "mail/new"
	MAIL_CONENT      = "mail/content"
	MAIL_CONENT_HTML = "mail/content_html"
)

type MailSearchOptions struct {
	Page     int
	PageSize int
	OrderBy  string
	TplName  string
	Type     int
	Bid      int64
	Keyword  string
}

// @Summary Get mail list
// @Description Get a list of mails with pagination and search options
// @Accept json
// @Produce html
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Param keyword query string false "Search keyword"
// @Success 200 {string} string "OK"
// @Failure 500 {object} string "Internal Server Error"
// @Router /mail [get]
func RenderMailSearch(c *context.Context, opts *MailSearchOptions) {
	var (
		mails []db.Mail
		count int64
		err   error
	)

	if opts.Page <= 0 {
		opts.Page = 1
	}

	if opts.PageSize <= 0 {
		opts.PageSize = conf.App.PageSize
	}

	mails, err = db.MailList(opts.Type, opts.Page, opts.PageSize, opts.Keyword)
	if err != nil {
		c.Error(err, "MailList")
		return
	}

	count, err = db.MailCountWithOpts(&db.MailSearchOptions{Type: opts.Type, Keyword: opts.Keyword})
	if err != nil {
		c.Error(err, "MailCountWithOpts")
		return
	}

	c.Data["Keyword"] = opts.Keyword
	c.Data["Mails"] = mails
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, opts.Page, c.QueryInt("page"))

	c.Success(opts.TplName)
}

func Flags(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.flags")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")
	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeFlags,

		Bid: bid,
	})
}

func Sent(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.sent")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeSend,
		Bid:      bid,
	})
}

func Draft(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.draft")
	c.Data["PageIsMail"] = true
	c.Data["PageIsMailDraft"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeDraft,
		Bid:      bid,
	})
}

func Deleted(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.deleted")
	c.Data["PageIsMail"] = true
	c.Data["PageIsMailDeleted"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeDeleted,
		Bid:      bid,
	})
}

func HardDeleteDraftMail(c *context.Context) {
	id := c.ParamsInt64(":id")

	mail, err := db.MailById(id)
	if err != nil {
		c.Flash.Error(c.Tr("mail.draft.deletion_fail"))
		c.Redirect("/mail/draft")
		return
	}

	err = db.MailHardDeleteById(mail.Uid, mail.Id)
	if err != nil {
		c.Flash.Error(c.Tr("mail.draft.deletion_fail"))
	} else {
		c.Flash.Success(c.Tr("mail.draft.deletion_success"))
	}
	c.Redirect("/mail/draft")
}

func Junk(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.junk")
	c.Data["PageIsMail"] = true

	bid := c.ParamsInt64(":bid")

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeJunk,
		Bid:      bid,
	})
}

func Mail(c *context.Context) {

	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMail"] = true

	// c.Success(MAIL)
	bid := c.ParamsInt64(":bid")
	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeInbox,
		Bid:      bid,
	})
}

func New(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	bid := c.ParamsInt64(":bid")
	id := c.ParamsInt64(":id")

	mail, err := db.MailById(id)
	if err != nil {
		c.Error(err, "MailById")
		return
	}
	content, err := db.MailContentRead(mail.Uid, mail.Id)
	if err != nil {
		c.Error(err, "MailContentRead")
		return
	}
	email, err := mcopa.Parse(bufio.NewReader(strings.NewReader(content)))
	if err != nil {
		c.Error(err, "mcopa.Parse")
		return
	}

	if strings.EqualFold(email.TextBody, "") {
		content = email.HTMLBody
	} else {
		content = email.TextBody
	}

	c.Data["Bid"] = bid
	c.Data["id"] = id

	c.Data["Mail"] = mail
	c.Data["MailContent"] = content

	c.Data["EditorLang"] = tools.ToEditorLang(c.Data["NowLang"].(string))

	c.Success(MAIL_NEW)
}

func NewPost(c *context.Context, f form.SendMail) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		c.RenderWithErr(c.Tr("mail.new.default_not_set"), MAIL_NEW, &f)
		return
	}

	mail_from := fmt.Sprintf("%s@%s", c.User.Name, from)
	tc, err := tmail.GetMailSend(mail_from, f.ToMail, f.Subject, f.Content)
	if err != nil {
		c.RenderWithErr(err.Error(), MAIL_NEW, &f)
		return
	}

	if f.Id != 0 {
		_, err = db.MailUpdate(f.Id, c.User.Id, 0, mail_from, f.ToMail, tc, 0, false)
	} else {
		_, err = db.MailPushSend(c.User.Id, mail_from, f.ToMail, tc, false)
	}

	if err != nil {
		c.RenderWithErr(err.Error(), MAIL_NEW, &f)
		return
	}

	c.RedirectSubpath("/mail/sent")
}

func NewPostDraft(c *context.Context, f form.SendMail) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	from, err := db.DomainGetMainForDomain()
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	mail_from := fmt.Sprintf("%s@%s", c.User.Name, from)
	tc, err := tmail.GetMailSend(mail_from, f.ToMail, f.Subject, f.Content)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	var mid int64
	if f.Id != 0 {
		mid, err = db.MailUpdate(f.Id, c.User.Id, 0, mail_from, f.ToMail, tc, 0, true)
	} else {
		mid, err = db.MailPushSend(c.User.Id, mail_from, f.ToMail, tc, true)
	}

	if err == nil {
		r := make(map[string]int64)
		r["id"] = mid

		c.OKDATA(c.Tr("common.success"), r)
		return
	}

	c.Fail(-1, c.Tr("common.fail"))
}

func Content(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	r, err := db.MailById(id)
	if err != nil {
		c.Error(err, "MailById")
		return
	}
	c.Data["Mail"] = r

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	c.Data["ParseMail"] = email

	c.Success(MAIL_CONENT)
}

func ContentHtml(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	r, err := db.MailById(id)
	if err != nil {
		c.Error(err, "MailById")
		return
	}
	c.Data["Mail"] = r

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}
	c.Data["ParseMail"] = email

	c.Success(MAIL_CONENT_HTML)
}

func ContentDownload(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	r, err := db.MailById(id)

	if err != nil {
		c.Error(err, "MailById")
		return
	}
	emailFilePath := db.MailContentFilename(r.Uid, id)
	tmpEmailName := fmt.Sprintf("imail_%d.eml", id)
	c.ServeFile(emailFilePath, tmpEmailName)
}

func ContentAttach(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsMailContent"] = true

	id := c.ParamsInt64(":id")
	c.Data["id"] = id

	aid := c.ParamsInt(":aid")

	bid := c.ParamsInt64(":bid")
	c.Data["Bid"] = bid

	r, err := db.MailById(id)
	if err != nil {
		c.Error(err, "MailById")
		return
	}
	c.Data["Mail"] = r

	contentData, err := db.MailContentRead(r.Uid, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	content := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(content)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}
	c.Data["ParseMail"] = email

	attachFile, err := ioutil.ReadAll(email.Attachments[aid].Data)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}
	pathName := "/tmp/" + email.Attachments[aid].Filename
	err = tools.WriteFile(pathName, string(attachFile))
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	c.ServeFile(pathName, email.Attachments[aid].Filename)
	os.RemoveAll(pathName)

	// return macaron.ReturnStruct{Code: http.StatusOK, Data: string(attachFile)}
}

func ContentDemo(c *context.Context) {

	id := c.ParamsInt64(":id")
	contentData, err := db.MailContentRead(1, id)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	bufferedBody := bufio.NewReader(strings.NewReader(contentData))
	email, err := mcopa.Parse(bufferedBody)
	if err != nil {
		c.Fail(-1, err.Error())
		return
	}

	c.OKDATA("ok", email)
}

/****************************************************
 * API for web frontend call
 ****************************************************/
func ApiDeleted(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	err = db.MailSoftDeleteByIds(idsSlice)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

// TODO:硬删除
func ApiHardDeleted(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	err = db.MailHardDeleteByIds(idsSlice)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

func ApiRead(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	err = db.MailSeenByIds(idsSlice)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

func ApiUnread(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	err = db.MailUnSeenByIds(idsSlice)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

func ApiStar(c *context.Context, f form.MailIDs) {
	int64ID, err := strconv.ParseInt(f.Ids, 10, 64)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}
	err = db.MailSetFlagsById(int64ID, 1)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

func ApiUnStar(c *context.Context, f form.MailIDs) {
	int64ID, err := strconv.ParseInt(f.Ids, 10, 64)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}
	err = db.MailSetFlagsById(int64ID, 0)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
	} else {
		c.OK(c.Tr("common.success"))
	}
}

func ApiMove(c *context.Context, f form.MailIDs) {
	ids := f.Ids
	dir := f.Dir

	idsSlice, err := tools.ToSlice(ids)
	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	if strings.EqualFold(dir, "deleted") {
		err = db.MailSoftDeleteByIds(idsSlice)
		if err != nil {
			c.Fail(-1, c.Tr("common.fail"))
			return
		}
		c.OK(c.Tr("common.success"))
		return
	}

	if strings.EqualFold(dir, "junk") {
		err = db.MailSetJunkByIds(idsSlice, 1)
		if err != nil {
			c.Fail(-1, c.Tr("common.fail"))
			return
		}
		c.OK(c.Tr("common.success"))
		return
	}

	c.Fail(-1, c.Tr("common.fail"))
	return
}
