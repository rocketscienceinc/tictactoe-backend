package websocket

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

func (that *Server) sendMessage(bufrw bufio.ReadWriter, action string, payload ResponsePayload) error {
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

	if err := writeFrame(bufrw, f); err != nil {
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

func writeFrame(bufrw bufio.ReadWriter, frameData frame) error {
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

	if err := bufrw.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

func (that *Server) readRequest(bufrw *bufio.ReadWriter) ([]byte, error) {
	header, err := that.readHeader(*bufrw)
	if err != nil {
		return nil, err
	}

	payload, err := that.readPayload(bufrw, header)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (that *Server) readHeader(bufrw bufio.ReadWriter) ([]byte, error) {
	header := make([]byte, 2)
	_, err := bufrw.Read(header)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	return header, nil
}

func (that *Server) readPayload(bufrw *bufio.ReadWriter, header []byte) ([]byte, error) {
	finBit := header[0] >> 7
	opCode := header[0] & 0x0f
	maskBit := header[1] >> 7
	payloadLen := header[1] & 0x7f

	size, _, err := that.readPayloadLength(*bufrw, payloadLen)
	if err != nil {
		return nil, err
	}

	mask, err := that.readMask(*bufrw, maskBit)
	if err != nil {
		return nil, err
	}

	payload, err := that.readData(*bufrw, size, mask)
	if err != nil {
		return nil, err
	}

	if finBit == 1 || opCode == 8 {
		return payload, nil
	}

	return nil, nil
}

func (that *Server) readPayloadLength(bufrw bufio.ReadWriter, payloadLen byte) (uint64, int, error) {
	if payloadLen < 126 {
		return uint64(payloadLen), 0, nil
	}

	if payloadLen == 126 {
		length := make([]byte, 2)
		_, err := bufrw.Read(length)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to read payload length: %w", err)
		}
		return uint64(binary.BigEndian.Uint16(length)), 2, nil
	}

	length := make([]byte, 8)
	_, err := bufrw.Read(length)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read payload length: %w", err)
	}

	return binary.BigEndian.Uint64(length), 8, nil
}

func (that *Server) readMask(bufrw bufio.ReadWriter, maskBit byte) ([]byte, error) {
	if maskBit == 0 {
		return nil, nil
	}

	mask := make([]byte, 4)
	_, err := bufrw.Read(mask)
	if err != nil {
		return nil, fmt.Errorf("failed to read mask: %w", err)
	}

	return mask, nil
}

func (that *Server) readData(bufrw bufio.ReadWriter, size uint64, mask []byte) ([]byte, error) {
	payload := make([]byte, size)
	_, err := io.ReadFull(bufrw, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}

	if mask != nil {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}

	return payload, nil
}
