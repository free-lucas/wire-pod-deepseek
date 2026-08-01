package wirepod_ttr

import (
	"bytes"
	"encoding/binary"
)



// CreateWavData 将音频数据转换为WAV格式的字节切片
func CreateWavData(audioData []float64, sampleRate, channels, bitsPerSample int) ([]byte, error) {
	var buf bytes.Buffer

	// 音频数据大小
	dataSize := len(audioData) * (bitsPerSample / 8)

	// 写入WAV文件头
	err := writeWavHeader(&buf, sampleRate, channels, bitsPerSample, dataSize)
	if err != nil {
		return nil, err
	}

	// 写入音频数据
	for _, sample := range audioData {
		// 将浮点数转换为16位PCM值
		pcmValue := int16(sample * 32767) // 32767是16位PCM的最大值
		err := binary.Write(&buf, binary.LittleEndian, pcmValue)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// 写入WAV文件头
func writeWavHeader(buf *bytes.Buffer, sampleRate, channels, bitsPerSample, dataSize int) error {
	// WAV文件头的常数
	riffHeaderSize := 36
	wavHeaderSize := riffHeaderSize + 8 + dataSize

	// RIFF头
	if err := binary.Write(buf, binary.LittleEndian, []byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(wavHeaderSize)); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, []byte("WAVE")); err != nil {
		return err
	}

	// fmt块
	if err := binary.Write(buf, binary.LittleEndian, []byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint16(channels)); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(sampleRate)); err != nil {
		return err
	}
	byteRate := sampleRate * channels * (bitsPerSample / 8)
	if err := binary.Write(buf, binary.LittleEndian, uint32(byteRate)); err != nil {
		return err
	}
	blockAlign := uint16(channels * (bitsPerSample / 8))
	if err := binary.Write(buf, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample)); err != nil {
		return err
	}

	// 数据块
	if err := binary.Write(buf, binary.LittleEndian, []byte("data")); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, uint32(dataSize)); err != nil {
		return err
	}

	return nil
}
