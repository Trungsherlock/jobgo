package h1b

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"

	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/google/uuid"
)

func NormalizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, suffix := range []string{", inc", ", inc", ", llc", ", ltd", "inc.", " inc", " llc", " ltd", " corp.", " corp", " co." } {
		name = strings.TrimSuffix(name, suffix)
	}

	name = strings.Map(func(r rune) rune {
		if r == ',' || r == '.' || r == '\'' {
			return -1
		}
		return r
	}, name)
	return strings.TrimSpace(name)
}

func ImportSponsors(db *database.DB, filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Decode UTF-16 LE (USCIS files use this encoding)
	utf16Decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	utf8Reader := transform.NewReader(f, utf16Decoder.NewDecoder())

	reader := csv.NewReader(utf8Reader)
	reader.Comma = '\t'
	reader.LazyQuotes = true

	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("reading header: %w", err)
	}

	colIdx := make(map[string]int)
	for i, col := range header {
		colIdx[strings.TrimSpace(strings.ToLower(col))] = i
	}

	count := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		employer := getCol(record, colIdx, "employer (petitioner) name")
		if employer == "" {
			continue
		}

		initApproval := atoi(getCol(record, colIdx, "new employment approval"))
		initDenial := atoi(getCol(record, colIdx, "new employment denial"))
		contApproval := atoi(getCol(record, colIdx, "continuation approval"))
		contDenial := atoi(getCol(record, colIdx, "continuation denial"))
		total := initApproval + initDenial + contApproval + contDenial

		approvalRate := 0.0
		if total > 0 {
			approvalRate = float64(initApproval+contApproval) / float64(total) * 100
		}

		sponsor := database.H1bSponsors{
			ID:                  uuid.New().String(),
            CompanyName:         employer,
            NormalizedName:      NormalizeName(employer),
            City:                getCol(record, colIdx, "petitioner city"),
            State:               getCol(record, colIdx, "petitioner state"),
            NaicsCode:           getCol(record, colIdx, "industry (naics) code"),
            FiscalYear:          atoi(getCol(record, colIdx, "fiscal year")),
            InitialApprovals:    initApproval,
            InitialDenials:      initDenial,
            ContinuingApprovals: contApproval,
            ContinuingDenials:   contDenial,
            ApprovalRate:        approvalRate,
            TotalPetitions:      total,
		}

		if err := db.UpsertH1bSponsor(sponsor); err != nil {
			continue
		}
		count++
	}
	return count, nil
}


func getCol(record []string, colIdx map[string]int, name string) string {
	if idx, ok := colIdx[name]; ok && idx < len(record) {
		return strings.TrimSpace(record[idx])
	}
	return ""
}

func atoi(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	// Handle decimal values like "1.0"
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(math.Round(f))
	}
	n, _ := strconv.Atoi(s)
	return n
}

// LinkCompanies matches tracked companies to H1B sponsor records
func LinkCompanies(db *database.DB) (int, error) {
    companies, err := db.ListCompanies()
    if err != nil {
        return 0, err
    }

    linked := 0
    for _, c := range companies {
        normalized := NormalizeName(c.Name)
        sponsor, err := db.FindSponsorByName(normalized)
        if err != nil {
            continue
        }
        if err := db.LinkCompanyToSponsor(c.ID, sponsor.ID, sponsor.ApprovalRate, sponsor.TotalPetitions); err != nil {
            continue
        }
        linked++
    }
    return linked, nil
}

