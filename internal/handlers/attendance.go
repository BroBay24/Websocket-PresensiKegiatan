package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BroBay24/WebsocketUTS/internal/models"
	wsHub "github.com/BroBay24/WebsocketUTS/internal/websocket"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type AttendanceHandler struct {
	db  *gorm.DB
	hub *wsHub.Hub
}

func NewAttendanceHandler(db *gorm.DB, hub *wsHub.Hub) *AttendanceHandler {
	return &AttendanceHandler{db: db, hub: hub}
}

func (h *AttendanceHandler) List(c *gin.Context) {
	var records []models.Attendance
	if err := h.db.Order("waktu DESC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengambil data presensi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":   records,
		"total":  len(records),
	})
}

func (h *AttendanceHandler) ExportExcel(c *gin.Context) {
	var records []models.Attendance
	if err := h.db.Order("waktu ASC").Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal menyiapkan data ekspor",
		})
		return
	}

	file := excelize.NewFile()
	sheet := file.GetSheetName(file.GetActiveSheetIndex())

	headers := []string{"No", "Nama", "NIM", "Jurusan", "Angkatan", "Waktu"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = file.SetCellValue(sheet, cell, header)
	}

	for index, record := range records {
		row := index + 2
		_ = file.SetCellValue(sheet, fmt.Sprintf("A%d", row), index+1)
		_ = file.SetCellValue(sheet, fmt.Sprintf("B%d", row), record.Nama)
		_ = file.SetCellValue(sheet, fmt.Sprintf("C%d", row), record.Nim)
		_ = file.SetCellValue(sheet, fmt.Sprintf("D%d", row), record.Jurusan)
		_ = file.SetCellValue(sheet, fmt.Sprintf("E%d", row), record.Angkatan)
		_ = file.SetCellValue(sheet, fmt.Sprintf("F%d", row), record.Waktu.Format("2006-01-02 15:04:05"))
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Gagal mengekspor data presensi",
		})
		return
	}

	filename := fmt.Sprintf("rekap-presensi-%s.xlsx", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	_, _ = c.Writer.Write(buffer.Bytes())
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type submitMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type submitPayload struct {
	Nama     string `json:"nama"`
	Nim      string `json:"nim"`
	Jurusan  string `json:"jurusan"`
	Angkatan string `json:"angkatan"`
}

func (h *AttendanceHandler) HandleWebsocket(c *gin.Context) {
	role := c.Query("role")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	switch role {
	case "dashboard":
		h.handleDashboardSocket(conn)
	default:
		h.handleFormSocket(conn)
	}
}

func (h *AttendanceHandler) handleDashboardSocket(conn *ws.Conn) {
	client := wsHub.NewClient(h.hub, conn)
	h.hub.Register(client)

	// kirim data awal
	go func() {
		var records []models.Attendance
		if err := h.db.Order("waktu DESC").Find(&records).Error; err == nil {
			payload := struct {
				Type    string                 `json:"type"`
				Payload map[string]any         `json:"payload"`
			}{
				Type: "attendance:init",
				Payload: map[string]any{
					"data":  records,
					"total": len(records),
				},
			}
			if data, err := json.Marshal(payload); err == nil {
				client.Send(data)
			}
		}
	}()

	go client.WritePump()
	go client.ReadPump()
}

func (h *AttendanceHandler) handleFormSocket(conn *ws.Conn) {
	defer conn.Close()

	for {
		var message submitMessage
		if err := conn.ReadJSON(&message); err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				logError("membaca pesan form", err)
			}
			return
		}

	if message.Type != "attendance:submit" {
		writeJSON(conn, map[string]any{
			"type": "attendance:ack",
			"payload": map[string]any{
				"success": false,
				"message": "Tipe pesan tidak dikenal",
			},
		})
		continue
	}

	var payload submitPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		writeJSON(conn, map[string]any{
			"type": "attendance:ack",
			"payload": map[string]any{
				"success": false,
				"message": "Format data tidak valid",
			},
		})
		continue
	}

	if resp := h.processSubmission(payload); resp != nil {
		writeJSON(conn, resp)
	}
	}
}

func (h *AttendanceHandler) processSubmission(payload submitPayload) map[string]any {
	nama := strings.TrimSpace(payload.Nama)
	nim := strings.TrimSpace(payload.Nim)
	jurusan := strings.TrimSpace(payload.Jurusan)
	angkatanNum, err := strconv.Atoi(strings.TrimSpace(payload.Angkatan))
	if err != nil {
		return buildAck(false, "Angkatan harus berupa angka")
	}

	if nama == "" || nim == "" || jurusan == "" {
		return buildAck(false, "Semua field wajib diisi")
	}

	var existing models.Attendance
	err = h.db.Where("nim = ?", nim).First(&existing).Error
	if err == nil {
		return buildAck(false, "NIM tersebut sudah tercatat hadir")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logError("memeriksa NIM", err)
		return buildAck(false, "Terjadi kesalahan pada server")
	}

	record := models.Attendance{
		Nama:     nama,
		Nim:      nim,
		Jurusan:  jurusan,
		Angkatan: angkatanNum,
	}

	if err := h.db.Create(&record).Error; err != nil {
		logError("menyimpan presensi", err)
		return buildAck(false, "Gagal menyimpan data presensi")
	}

	var total int64
	if err := h.db.Model(&models.Attendance{}).Count(&total).Error; err != nil {
		logError("menghitung total presensi", err)
	}

	h.hub.BroadcastAttendance(record, total)

	return buildAck(true, "Presensi berhasil dicatat. Terima kasih!")
}

func buildAck(success bool, message string) map[string]any {
	return map[string]any{
		"type": "attendance:ack",
		"payload": map[string]any{
			"success": success,
			"message": message,
		},
	}
}

func writeJSON(conn *ws.Conn, value any) {
	if err := conn.WriteJSON(value); err != nil {
		logError("mengirim pesan websocket", err)
	}
}

func logError(context string, err error) {
	fmt.Printf("error %s: %v\n", context, err)
}
