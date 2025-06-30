package admin

import (
	"fmt"
	"net"
	"strings"
	// "errors"

	"github.com/kelvinzer0/imail-ipv6/internal/app/context"
	"github.com/kelvinzer0/imail-ipv6/internal/app/form"
	"github.com/kelvinzer0/imail-ipv6/internal/conf"
	"github.com/kelvinzer0/imail-ipv6/internal/db"
	"github.com/kelvinzer0/imail-ipv6/internal/tools"
	"github.com/kelvinzer0/imail-ipv6/internal/tools/dkim"
)

const (
	DOMAIN     = "admin/domain/list"
	DOMAIN_NEW = "admin/domain/new"
)

func Domain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	d, _ := db.DomainList(1, 10)

	c.Data["Total"] = db.DomainCount()
	c.Data["Domain"] = d

	c.Success(DOMAIN)
}

func NewDomain(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true

	c.Success(DOMAIN_NEW)
}

func NewDomainPost(c *context.Context, f form.AdminCreateDomain) {
	c.Data["Title"] = c.Tr("admin.domain")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminDomain"] = true
	count := db.DomainCount()

	limit := 9
	if int(count) >= limit {
		c.FormErr("Domain")
		c.RenderWithErr(c.Tr("form.domain_add_limit_exceeded", limit), DOMAIN_NEW, &f)
		return
	}

	if c.HasError() {
		c.Success(DOMAIN_NEW)
		return
	}

	d := &db.Domain{
		Domain: f.Domain,
	}

	err := db.DomainCreate(d)
	if err != nil {
		c.FormErr("Domain")
		c.RenderWithErr(c.Tr("admin.domain.add_fail", f.Domain), DOMAIN_NEW, &f)
		return
	}

	c.Flash.Success(c.Tr("admin.domain.add_success", f.Domain))
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}

func DeleteDomain(c *context.Context) {
	id := c.ParamsInt64(":id")
	err := db.DomainDeleteById(id)
	if err != nil {
		c.Flash.Success(c.Tr("admin.domain.deletion_fail"))
	} else {
		c.Flash.Success(c.Tr("admin.domain.deletion_success"))
	}
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}

func InfoDomain(c *context.Context) {
	domain := c.Params(":domain")

	dataDir := conf.Web.Subpath + conf.Web.AppDataPath
	content, err := dkim.GetDomainDkimVal(dataDir, domain)

	if err != nil {
		c.Fail(-1, c.Tr("common.fail"))
		return
	}

	var d = make(map[string]string)

	localIp, _ := tools.GetPublicIP()
	d["ip"] = localIp
	d["dkim"] = content
	c.OKDATA("ok", d)
}

func CheckDomain(c *context.Context) {
	id := c.ParamsInt64(":id")
	d, err := db.DomainGetById(id)
	if err != nil {
		c.Flash.Error(c.Tr("common.fail"))
		c.Redirect(conf.Web.Subpath + "/admin/domain")
		return
	}
	domain := d.Domain

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		c.Flash.Error(c.Tr("admin.domain.check_fail", domain, err.Error()))
		c.Redirect(conf.Web.Subpath + "/admin/domain")
		return
	}

	// MX Record Check
	if len(mxRecords) > 0 && strings.Contains(mxRecords[0].Host, ".") {
		d.Mx = true
	} else {
		d.Mx = false
	}

	// A Record Check
	if d.Mx {
		host := strings.Trim(mxRecords[0].Host, ".")
		err = dkim.CheckDomainARecord(host)
		if err == nil {
			d.A = true
		} else {
			d.A = false
		}
	} else {
		d.A = false
	}

	// AAAA Record Check
	if d.Mx {
		host := strings.Trim(mxRecords[0].Host, ".")
		err = dkim.CheckDomainAAAARecord(host)
		if err == nil {
			d.AAAA = true
		} else {
			d.AAAA = false
		}
	} else {
		d.AAAA = false
	}

	// DMARC Check
	dmarcRecord, err := net.LookupTXT(fmt.Sprintf("_dmarc.%s", domain))
	if err != nil {
		// Log the error but don't fail the entire check
		fmt.Printf("Error looking up DMARC record for %s: %v\n", domain, err)
	}
	if len(dmarcRecord) > 0 {
		for _, rec := range dmarcRecord {
			if strings.Contains(strings.ToLower(rec), "v=dmarc1") {
				d.Dmarc = true
				break
			}
		}
	} else {
		d.Dmarc = false
	}

	// SPF Check
	spfRecord, err := net.LookupTXT(domain)
	if err != nil {
		// Log the error but don't fail the entire check
		fmt.Printf("Error looking up SPF record for %s: %v\n", domain, err)
	}
	if len(spfRecord) > 0 {
		for _, rec := range spfRecord {
			if strings.Contains(strings.ToLower(rec), "v=spf1") {
				d.Spf = true
				break
			}
		}
	} else {
		d.Spf = false
	}

	// DKIM Check
	dataDir := conf.Web.Subpath + conf.Web.AppDataPath
	dkimRecord, err := net.LookupTXT(fmt.Sprintf("default._domainkey.%s", domain))
	if err != nil {
		// Log the error but don't fail the entire check
		fmt.Printf("Error looking up DKIM record for %s: %v\n", domain, err)
	}
	if len(dkimRecord) > 0 {
		dkimContent, err := dkim.GetDomainDkimVal(dataDir, domain)
		if err != nil {
			fmt.Printf("Error getting DKIM value for %s: %v\n", domain, err)
		} else {
			for _, rec := range dkimRecord {
				if strings.EqualFold(dkimContent, rec) {
					d.Dkim = true
					break
				}
			}
		}
	} else {
		d.Dkim = false
	}

	err = db.DomainUpdateById(id, d)
	if err != nil {
		c.Flash.Error(c.Tr("admin.domain.update_fail", d.Domain, err.Error()))
	} else {
		c.Flash.Success(c.Tr("admin.domain.check_success", d.Domain))
	}
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}

func SetDefaultDomain(c *context.Context) {
	id := c.ParamsInt64(":id")

	err := db.DomainSetDefaultOnlyOne(id)
	if err != nil {
		c.Flash.Error(c.Tr("admin.domain.set_default_fail"))
	} else {
		c.Flash.Success(c.Tr("admin.domain.set_default_success"))
	}
	c.Redirect(conf.Web.Subpath + "/admin/domain")
}
