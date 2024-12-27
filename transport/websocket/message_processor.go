package websocket

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/rocketscienceinc/tictactoe-backend/internal/entity"
)

var (
	ErrUnsupportedOpcode              = errors.New("unsupported opcode")
	ErrFragmentedMessagesNotSupported = errors.New("fragmented messages are not supported")
)

// frame represents a WebSocket frame and its metadata.
type frame struct {
	isFin   bool   // Указывает, является ли этот фрейм последним в сообщении
	opCode  byte   // Код операции, указывающий тип данных (например, текстовое сообщение, бинарные данные и т.д.)
	length  uint64 // Длина полезной нагрузки (payload) фрейма
	payload []byte // Данные, передаваемые в фрейме
}

// Message represents a WebSocket message with an action type and a payload.
type Message struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Payload struct {
	Player  *entity.Player `json:"player,omitempty"`
	Game    *entity.Game   `json:"game,omitempty"`
	Error   string         `json:"error,omitempty"`
	Cell    *int           `json:"cell,omitempty"`
	Answer  string         `json:"answer,omitempty"`
	Message string         `json:"message,omitempty"`
}

func (that *Server) sendMessage(bufrw *bufio.ReadWriter, action string, payload Payload) error {
	response := Message{
		Action:  action,
		Payload: json.RawMessage(mustMarshal(payload)),
	}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	f := frame{
		isFin:   true,
		opCode:  1, // текстовое сообщение
		length:  uint64(len(responseBytes)),
		payload: responseBytes,
	}

	if err = writeFrame(bufrw, f); err != nil {
		return fmt.Errorf("failed to write frame: %w", err)
	}

	return nil
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func writeFrame(bufrw *bufio.ReadWriter, frameData frame) error {
	buf := make([]byte, 2)
	buf[0] |= frameData.opCode

	if frameData.isFin {
		buf[0] |= 0x80
	}

	switch {
	case frameData.length < 126:
		buf[1] |= byte(frameData.length)
	case frameData.length < 1<<16:
		buf[1] |= 126
		size := make([]byte, 2)
		binary.BigEndian.PutUint16(size, uint16(frameData.length))
		buf = append(buf, size...) //nolint: makezero // idk how to rewrite this
	default:
		buf[1] |= 127
		size := make([]byte, 8)
		binary.BigEndian.PutUint64(size, frameData.length)
		buf = append(buf, size...) //nolint: makezero // idk how to rewrite this
	}

	buf = append(buf, frameData.payload...) //nolint: makezero // idk how to rewrite this

	_, err := bufrw.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to write frame: %w", err)
	}

	if err = bufrw.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

func (that *Server) readRequest(bufrw *bufio.ReadWriter) ([]byte, error) {
	header, err := readHeader(bufrw)
	if err != nil {
		return nil, err
	}

	payload, err := readPayload(bufrw, header)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func readHeader(bufrw *bufio.ReadWriter) ([]byte, error) {
	header := make([]byte, 2)
	_, err := bufrw.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	return header, nil
}

func readPayload(bufrw *bufio.ReadWriter, header []byte) ([]byte, error) {
	fin := (header[0] & 0x80) != 0
	opcode := header[0] & 0x0F
	mask := (header[1] & 0x80) != 0
	payloadLen := uint64(header[1] & 0x7F)

	// Чтение расширенной длины полезной нагрузки
	if payloadLen == 126 {
		extended := make([]byte, 2)
		_, err := io.ReadFull(bufrw, extended)
		if err != nil {
			return nil, fmt.Errorf("failed to read extended payload length: %w", err)
		}
		payloadLen = uint64(binary.BigEndian.Uint16(extended))
	} else if payloadLen == 127 {
		extended := make([]byte, 8)
		_, err := io.ReadFull(bufrw, extended)
		if err != nil {
			return nil, fmt.Errorf("failed to read extended payload length: %w", err)
		}
		payloadLen = binary.BigEndian.Uint64(extended)
	}

	// Чтение маскирующего ключа
	var maskingKey []byte
	if mask {
		maskingKey = make([]byte, 4)
		_, err := io.ReadFull(bufrw, maskingKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read masking key: %w", err)
		}
	}

	// Чтение полезной нагрузки
	payload := make([]byte, payloadLen)
	_, err := io.ReadFull(bufrw, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	// Применение маски
	if mask {
		for i := range payload {
			payload[i] ^= maskingKey[i%4]
		}
	}

	// Обработка опкодов
	if opcode == 8 {
		// Фрейм закрытия соединения
		return nil, io.EOF
	} else if opcode != 1 {
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedOpcode, opcode)
	}

	if !fin {
		// Обработка фрагментированных сообщений, если необходимо
		return nil, ErrFragmentedMessagesNotSupported
	}

	return payload, nil
}
