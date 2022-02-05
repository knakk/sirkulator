package sirkulator

import (
	"testing"

	"golang.org/x/text/language"
)

func TestYearRange(t *testing.T) {
	tests := []struct {
		input  YearRange
		want   string
		wantEn string
		wantNo string
	}{
		{
			input:  YearRange{From: 1981},
			want:   "1981–",
			wantEn: "1981–",
			wantNo: "1981–",
		},
		{
			input:  YearRange{To: 1981},
			want:   "?–1981",
			wantEn: "?–1981",
			wantNo: "?–1981",
		},
		{
			input:  YearRange{From: 1906, To: 1989},
			want:   "1906–1989",
			wantEn: "1906–1989",
			wantNo: "1906–1989",
		},
		{
			input:  YearRange{From: 1870, To: 1923, Approx: true},
			want:   "ca. 1870–1923",
			wantEn: "ca. 1870–1923",
			wantNo: "ca. 1870–1923",
		},
		{
			input:  YearRange{To: 1923, Approx: true},
			want:   "ca. ?–1923",
			wantEn: "ca. ?–1923",
			wantNo: "ca. ?–1923",
		},
		{
			input:  YearRange{From: 1870, Approx: true},
			want:   "ca. 1870–",
			wantEn: "ca. 1870–",
			wantNo: "ca. 1870–",
		},
		{
			input:  YearRange{From: 1700, To: 1800, Approx: true},
			want:   "ca. 1700–1800",
			wantEn: "18th century",
			wantNo: "1700-tallet",
		},
		{
			input:  YearRange{From: 1700, To: 1900, Approx: true},
			want:   "ca. 1700–1900",
			wantEn: "18/19th century",
			wantNo: "17/1800-tallet",
		},
		{
			input:  YearRange{From: -500, To: -400, Approx: true},
			want:   "ca. 500–400 BCE",
			wantEn: "6th century BCE",
			wantNo: "500-tallet f.Kr",
		},
		{
			input:  YearRange{From: -51, To: 21},
			want:   "51 BCE–21 AD",
			wantEn: "51 BCE–21 AD",
			wantNo: "51 f.Kr–21 e.Kr",
		},
		// TODO testcases:
		// YearRange{From: 1981, To: 1981}
		// YearRange{From: 1981, To: 1981, Approx:true}
		// YearRange{From: -200, To: 300, Approx:true}
	}

	for _, test := range tests {
		if got := test.input.String(); got != test.want {
			t.Errorf("got %q; want %q", got, test.want)
		}
		if got := test.input.Label(language.English); got != test.wantEn {
			t.Errorf("got %q; want %q", got, test.wantEn)
		}
		if got := test.input.Label(language.Norwegian); got != test.wantNo {
			t.Errorf("got %q; want %q", got, test.wantNo)
		}
	}
}
