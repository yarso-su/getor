package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yarso-su/getor/internal/pdf"
	"github.com/yarso-su/nera"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, nera.ErrTypeMismatch) || errors.Is(err, nera.ErrIndexOutOfRange) {
			fmt.Fprintf(os.Stderr, "getor: file does not match the expected receipt shape: %v\n\n", err)
			fmt.Fprintln(os.Stderr, "See the format reference: https://github.com/yarso-su/getor\n")
		} else {
			fmt.Fprintf(os.Stderr, "getor: %v\n", err)
		}
		os.Exit(1)
	}
}

func run() error {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "getor: generator of pdf receipts from .nera files")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "See https://github.com/yarso-su/getor for input format.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  getor <input.nera>")
		fmt.Fprintln(os.Stderr)
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	file, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s.pdf", strings.TrimSuffix(filepath.Base(flag.Arg(0)), filepath.Ext(flag.Arg(0))))

	data, err := nera.Parse(string(file))
	if err != nil {
		return err
	}

	header, err := data.At(0)
	if err != nil {
		return err
	}

	// Support for single and multiple concepts
	var concepts nera.LiteralGroupCollection
	concepts, err = data.AtGroupCollection(4)
	if err != nil {
		if errors.Is(err, nera.ErrTypeMismatch) {
			concept, err := data.AtGroup(4)
			if err != nil {
				return err
			}

			values := make([][]string, 1)
			values[0] = concept.Values

			concepts.Keys = concept.Keys
			concepts.Values = values
		} else {
			return err
		}
	}

	note, err := data.At(7)
	if err != nil {
		return err
	}

	// Measure the note text height and
	// add it to the document height
	measure := pdf.NewFile(0)
	measure.SetFont(pdf.Font, "", pdf.TextSize)
	noteHeight := float64(len(measure.SplitText(
		note.Value,
		pdf.Width-pdf.Margin*2,
	)) * pdf.TextSize)
	measure.Close()

	height := 260 + float64((len(concepts.Values)*(pdf.TextSize+3))+155) + noteHeight
	doc := pdf.NewFile(height)
	doc.Ln(1)

	doc.SetFont(pdf.Font, "B", pdf.SubtitleSize)
	doc.CellFormat(0, pdf.SubtitleSize, header.Key, "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, header.Value, "", 1, "L", false, 0, "")
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap*2)

	dates, err := data.AtGroup(1)
	if err != nil {
		return err
	}

	half := float64((pdf.Width / 2) - pdf.Margin)
	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(half, pdf.TextSize, dates.Keys[0], "", 0, "L", false, 0, "")
	doc.CellFormat(half, pdf.TextSize, dates.Keys[1], "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.CellFormat(half, pdf.TextSize, dates.Values[0], "", 0, "L", false, 0, "")
	doc.CellFormat(half, pdf.TextSize, dates.Values[1], "", 1, "L", false, 0, "")
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap)

	issuer, err := data.AtGroup(2)
	if err != nil {
		return err
	}

	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, issuer.Keys[0], "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, issuer.Values[0], "", 1, "L", false, 0, "")
	doc.Ln(2)
	doc.CellFormat(0, pdf.TextSize, issuer.Values[1], "", 1, "L", false, 0, "")
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap)

	payer, err := data.AtGroup(3)
	if err != nil {
		return err
	}

	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, payer.Keys[0], "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, payer.Values[0], "", 1, "L", false, 0, "")
	doc.Ln(2)
	doc.CellFormat(0, pdf.TextSize, payer.Values[1], "", 1, "L", false, 0, "")
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap*3)

	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(half, pdf.TextSize, concepts.Keys[0], "", 0, "L", false, 0, "")
	doc.CellFormat(half/6*2, pdf.TextSize, concepts.Keys[1], "", 0, "R", false, 0, "")
	doc.CellFormat(half/6*2, pdf.TextSize, concepts.Keys[2], "", 0, "R", false, 0, "")
	doc.CellFormat(half/6*2, pdf.TextSize, concepts.Keys[3], "", 1, "R", false, 0, "")
	doc.Ln(6)

	doc.SetAlpha(0.95, "Normal")
	for _, concept := range concepts.Values {
		doc.SetFont(pdf.Font, "", pdf.TextSize)
		doc.CellFormat(half, pdf.TextSize, concept[0], "", 0, "L", false, 0, "")
		doc.CellFormat(half/6*2, pdf.TextSize, concept[1], "", 0, "R", false, 0, "")
		doc.CellFormat(half/6*2, pdf.TextSize, concept[2], "", 0, "R", false, 0, "")
		doc.CellFormat(half/6*2, pdf.TextSize, concept[3], "", 1, "R", false, 0, "")
		doc.Ln(3)
	}
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap)

	total, err := data.At(5)
	if err != nil {
		return err
	}

	doc.SetFont(pdf.Font, "B", pdf.SubtitleSize)
	doc.CellFormat(0, pdf.SubtitleSize, fmt.Sprintf("%s %s", total.Key, total.Value), "", 1, "R", false, 0, "")
	doc.Ln(1)

	doc.SetY(doc.GetY() + pdf.Gap*2)

	paymentMethod, err := data.At(6)
	if err != nil {
		return err
	}

	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, paymentMethod.Key, "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, paymentMethod.Value, "", 1, "L", false, 0, "")
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	doc.SetY(doc.GetY() + pdf.Gap)

	doc.SetFont(pdf.Font, "B", pdf.TextSize)
	doc.CellFormat(0, pdf.TextSize, note.Key, "", 1, "L", false, 0, "")
	doc.Ln(3)
	doc.SetAlpha(0.95, "Normal")
	doc.SetFont(pdf.Font, "", pdf.TextSize)
	doc.MultiCell(0, pdf.TextSize, note.Value, "", "L", false)
	doc.Ln(1)
	doc.SetAlpha(1, "Normal")

	var buf bytes.Buffer
	err = doc.Output(&buf)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
