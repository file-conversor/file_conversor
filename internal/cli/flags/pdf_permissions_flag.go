// internal/cli/flags/pdf_permissions.go

package flags

import "github.com/file-conversor/file_conversor/internal/core/pdf"

type PdfPermissionsFlag struct {
	Permissions []string `short:"m" optional:"" default:"all" enum:"all,print,modify,copy,annotate,none" help:"Permissions for encrypted PDF files (comma-separated). Options: ${enum}."`
}

func (c *PdfPermissionsFlag) Get() pdf.EncryptPermissions {
	perms := pdf.EncryptPermissions{}
	for _, perm := range c.Permissions {
		switch perm {
		case "print":
			perms.PermissionPrint = true
		case "modify":
			perms.PermissionModify = true
		case "copy":
			perms.PermissionExtract = true
		case "annotate":
			perms.PermissionModAnnFillForm = true
		case "none":
			// no permissions
		default:
			perms.PermissionAll = true
		}
	}
	return perms
}
