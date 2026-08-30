package pdf

import (
	"codeberg.org/go-pdf/fpdf"
	"github.com/yarso-su/getor/internal/assets"
)

const (
	Font   = "Lato"
	Width  = 470
	Margin = 15

	TitleSize    = 20
	SubtitleSize = 16
	TextSize     = 11
	Gap          = 14
)

func NewFile(height float64) *fpdf.Fpdf {
	doc := fpdf.NewCustom(&fpdf.InitType{
		UnitStr: "pt",
		Size: fpdf.SizeType{
			Wd: Width,
			Ht: height,
		},
	})

	doc.AddPage()
	doc.SetMargins(Margin, Margin, Margin)
	doc.SetAutoPageBreak(false, Margin)

	doc.AddUTF8FontFromBytes(Font, "", assets.LatoRegular)
	doc.AddUTF8FontFromBytes(Font, "B", assets.LatoBold)

	doc.SetLineWidth(20)
	doc.SetDrawColor(0, 0, 0)
	doc.Line(0, 0, Width, 0)
	doc.SetLineWidth(1)

	return doc
}
