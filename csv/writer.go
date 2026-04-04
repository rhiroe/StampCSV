package csv

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// 列インデックス
	colDay    = 0  // 日
	colTime   = 1  // 時間（合計）
	colFirst  = 2  // 開始/終了ペア列の先頭
	numPairs  = 5  // 開始/終了ペア数
	colBreak  = 12 // 休憩等
	totalCols = 14 // 総列数（末尾の空列含む）
)

// Stamp はCSVファイルへタイムスタンプを記録する。
// stampType: "in"=開始, "out"=終了
func Stamp(dir string, stampType string) error {
	now := time.Now()
	return stamp(dir, stampType, now)
}

func stamp(dir string, stampType string, now time.Time) error {
	filePath := filepath.Join(dir, now.Format("2006-01")+".csv")

	rows, err := readOrInit(filePath, now)
	if err != nil {
		return fmt.Errorf("CSV読み込みエラー: %w", err)
	}

	day := now.Day()

	// OUT かつ前日に未終了の区間がある場合は日付またぎとみなし
	// 終了時刻を 24+時 形式で前日行に記録する
	if stampType == "out" && day > 1 {
		prevIdx := dayRowIndex(rows, day-1)
		if prevIdx >= 0 && hasPendingSession(rows[prevIdx]) {
			overTimeStr := fmt.Sprintf("%d:%02d", now.Hour()+24, now.Minute())
			rows, err = writeStamp(rows, day-1, "out", overTimeStr)
			if err != nil {
				return err
			}
			rows = recalcSummary(rows)
			return writeCSV(filePath, rows)
		}
	}

	timeStr := fmt.Sprintf("%d:%02d", now.Hour(), now.Minute())
	rows, err = writeStamp(rows, day, stampType, timeStr)
	if err != nil {
		return err
	}

	rows = recalcSummary(rows)

	return writeCSV(filePath, rows)
}

// readOrInit はCSVを読み込む。ファイルがなければ初期行を作成する。
func readOrInit(filePath string, now time.Time) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return initRows(now), nil
		}
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

// initRows は月の初期行（ヘッダー + 1〜lastDay日 + 集計行）を生成する。
func initRows(now time.Time) [][]string {
	year, month, _ := now.Date()
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	header := []string{"日", "時間", "開始", "終了", "開始", "終了", "開始", "終了", "開始", "終了", "開始", "終了", "休憩等", ""}
	rows := make([][]string, 0, lastDay+4)
	rows = append(rows, header)
	for d := 1; d <= lastDay; d++ {
		row := makeEmptyRow(d)
		rows = append(rows, row)
	}
	// 時間計行
	rows = append(rows, makeFixedRow("時間計"))
	// 日数行
	rows = append(rows, makeFixedRow("日数"))
	// 時間修正枠行
	rows = append(rows, []string{"時間修正枠", "", ""})
	return rows
}

func makeEmptyRow(day int) []string {
	row := make([]string, totalCols)
	row[colDay] = strconv.Itoa(day)
	return row
}

func makeFixedRow(label string) []string {
	row := make([]string, totalCols)
	row[colDay] = label
	return row
}

// hasPendingSession は行に開始済み・未終了の区間があるかを返す。
func hasPendingSession(row []string) bool {
	for p := 0; p < numPairs; p++ {
		startCol := colFirst + p*2
		endCol := startCol + 1
		if startCol >= len(row) || endCol >= len(row) {
			break
		}
		if row[startCol] != "" && row[endCol] == "" {
			return true
		}
	}
	return false
}

// writeStamp は対象日行に stampType（"in"/"out"）を書き込む。
func writeStamp(rows [][]string, day int, stampType, timeStr string) ([][]string, error) {
	idx := dayRowIndex(rows, day)
	if idx < 0 {
		return rows, fmt.Errorf("%d日の行が見つかりません", day)
	}

	ensureCols(&rows[idx], colFirst+numPairs*2)

	switch stampType {
	case "in":
		// 次の空き「開始」列に書く
		for p := 0; p < numPairs; p++ {
			startCol := colFirst + p*2
			endCol := startCol + 1
			if rows[idx][startCol] == "" {
				// 直前のペアが終了済みか確認（p>0 の場合）
				if p > 0 && rows[idx][endCol-1] == "" {
					return rows, fmt.Errorf("前の区間が終了していません")
				}
				rows[idx][startCol] = timeStr
				return rows, nil
			}
		}
		return rows, fmt.Errorf("開始列がいっぱいです（最大%d区間）", numPairs)

	case "out":
		// 最後の開始済み・終了なしペアに書く
		for p := numPairs - 1; p >= 0; p-- {
			startCol := colFirst + p*2
			endCol := startCol + 1
			if rows[idx][startCol] != "" && rows[idx][endCol] == "" {
				rows[idx][endCol] = timeStr
				return rows, nil
			}
		}
		return rows, fmt.Errorf("対応する開始がありません")

	default:
		return rows, fmt.Errorf("不明なスタンプ種別: %s", stampType)
	}
}

// recalcSummary は全行の「時間」列、「時間計」「日数」行を再計算する。
func recalcSummary(rows [][]string) [][]string {
	totalMinutes := 0
	workDays := 0

	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		day, err := strconv.Atoi(row[colDay])
		if err != nil || day < 1 || day > 31 {
			continue
		}

		mins := calcDayMinutes(row)
		ensureCols(&rows[i], totalCols)
		if mins > 0 {
			rows[i][colTime] = minutesToHHMM(mins)
			totalMinutes += mins
			workDays++
		} else {
			rows[i][colTime] = ""
		}
	}

	// 時間計・日数行を更新
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		switch row[colDay] {
		case "時間計":
			ensureCols(&rows[i], totalCols)
			rows[i][colTime] = minutesToHHMM(totalMinutes)
		case "日数":
			ensureCols(&rows[i], totalCols)
			if workDays > 0 {
				rows[i][colTime] = strconv.Itoa(workDays)
			} else {
				rows[i][colTime] = ""
			}
		}
	}

	return rows
}

// calcDayMinutes は1行の全区間の合計分を返す。
func calcDayMinutes(row []string) int {
	total := 0
	for p := 0; p < numPairs; p++ {
		startCol := colFirst + p*2
		endCol := startCol + 1
		if startCol >= len(row) || endCol >= len(row) {
			break
		}
		start := row[startCol]
		end := row[endCol]
		if start == "" || end == "" {
			continue
		}
		sm, err1 := parseHHMM(start)
		em, err2 := parseHHMM(end)
		if err1 != nil || err2 != nil {
			continue
		}
		d := em - sm
		if d < 0 {
			d += 24 * 60
		}
		total += d
	}
	return total
}

func parseHHMM(s string) (int, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time: %s", s)
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid time: %s", s)
	}
	return h*60 + m, nil
}

func minutesToHHMM(mins int) string {
	h := int(math.Abs(float64(mins))) / 60
	m := int(math.Abs(float64(mins))) % 60
	if mins < 0 {
		return fmt.Sprintf("-%d:%02d", h, m)
	}
	return fmt.Sprintf("%d:%02d", h, m)
}

// dayRowIndex は rows の中で日付が day の行インデックスを返す。
func dayRowIndex(rows [][]string, day int) int {
	dayStr := strconv.Itoa(day)
	for i, row := range rows {
		if len(row) > 0 && row[colDay] == dayStr {
			return i
		}
	}
	return -1
}

func ensureCols(row *[]string, n int) {
	for len(*row) < n {
		*row = append(*row, "")
	}
}

// writeCSV はCSVファイルに書き込む。
func writeCSV(filePath string, rows [][]string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("CSV作成エラー: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for _, row := range rows {
		// 末尾の空フィールドを保持するためそのまま書く
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}
