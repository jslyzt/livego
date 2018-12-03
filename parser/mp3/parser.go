package mp3

import "errors"

// Parser 解析器
type Parser struct {
	samplingFrequency int
}

// NewParser 新解析器
func NewParser() *Parser {
	return &Parser{}
}

// sampling_frequency - indicates the sampling frequency, according to the following table.
// '00' 44.1 kHz
// '01' 48 kHz
// '10' 32 kHz
// '11' reserved
var mp3Rates = []int{44100, 48000, 32000}
var (
	errMp3DataInvalid = errors.New("mp3data  invalid")
	errIndexInvalid   = errors.New("invalid rate index")
)

// Parse 解析
func (parser *Parser) Parse(src []byte) error {
	if len(src) < 3 {
		return errMp3DataInvalid
	}
	index := (src[2] >> 2) & 0x3
	if index <= byte(len(mp3Rates)-1) {
		parser.samplingFrequency = mp3Rates[index]
		return nil
	}
	return errIndexInvalid
}

// SampleRate 简单码率
func (parser *Parser) SampleRate() int {
	if parser.samplingFrequency == 0 {
		parser.samplingFrequency = 44100
	}
	return parser.samplingFrequency
}
