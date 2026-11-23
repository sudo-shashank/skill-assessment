package report

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"

	"skill-assessment/go-service/internal/student"
)

func Generate(s *student.Student) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("Student Report", false)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 12, "Student Report")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC1123)))
	pdf.Ln(12)

	systemAccess := "No"
	if s.SystemAccess {
		systemAccess = "Yes"
	}
	roll := "-"
	if s.Roll != nil {
		roll = strconv.Itoa(*s.Roll)
	}

	writeSection(pdf, "Identity", [][2]string{
		{"Name", s.Name},
		{"Email", s.Email},
		{"Phone", optStr(s.Phone)},
		{"Gender", optStr(s.Gender)},
		{"Date of Birth", optStr(s.DOB)},
		{"System Access", systemAccess},
	})

	writeSection(pdf, "Academics", [][2]string{
		{"Class", optStr(s.Class)},
		{"Section", optStr(s.Section)},
		{"Roll", roll},
		{"Admission Date", optStr(s.AdmissionDate)},
		{"Reporter", optStr(s.ReporterName)},
	})

	writeSection(pdf, "Guardians", [][2]string{
		{"Father Name", optStr(s.FatherName)},
		{"Father Phone", optStr(s.FatherPhone)},
		{"Mother Name", optStr(s.MotherName)},
		{"Mother Phone", optStr(s.MotherPhone)},
		{"Guardian Name", optStr(s.GuardianName)},
		{"Guardian Phone", optStr(s.GuardianPhone)},
		{"Relation of Guardian", optStr(s.RelationOfGuardian)},
	})

	writeSection(pdf, "Addresses", [][2]string{
		{"Current Address", optStr(s.CurrentAddress)},
		{"Permanent Address", optStr(s.PermanentAddress)},
	})

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("build pdf: %w", err)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("write pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func writeSection(pdf *fpdf.Fpdf, title string, rows [][2]string) {
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, title)
	pdf.Ln(8)

	for _, row := range rows {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(55, 7, row[0]+":")
		pdf.SetFont("Arial", "", 11)
		value := row[1]
		if value == "" {
			value = "-"
		}
		pdf.MultiCell(0, 7, value, "", "", false)
	}
	pdf.Ln(4)
}

func optStr(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
