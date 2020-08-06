package tcpx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"io"
)

// tcpx's tool to help build expected stream for communicating
type Packx struct {
	//Marshaller Marshaller
}

//// a package scoped packx instance
//var packx = NewPackx(nil)
//var PackJSON = NewPackx(JsonMarshaller{})
//var PackTOML = NewPackx(TomlMarshaller{})
//var PackXML = NewPackx(XmlMarshaller{})
//var PackYAML = NewPackx(YamlMarshaller{})
//var PackProtobuf = NewPackx(ProtobufMarshaller{})

// New a packx instance, specific a marshaller for communication.
// If marshaller is nil, official jsonMarshaller is put to used.
func NewPackx() *Packx {
	return &Packx{}
}

// Pack src with specific messageID and optional headers
// Src has not been marshaled yet.Whatever you put as src, it will be marshaled by packx.Marshaller.
func (packx Packx) Pack(messageID int32, src []byte) ([]byte, error) {
	return PackWithMarshaller(Message{MessageID: messageID, Body: src})
}

// Unpack
// Stream is a block of length,messageID,headerLength,bodyLength,header,body.
// Dest refers to the body, it can be dynamic by messageID.
//
// Before use this, users should be aware of which struct used as `dest`.
// You can use stream's messageID for judgement like:
// messageID,_:= packx.MessageIDOf(stream)
// switch messageID {
//     case 1:
//       packx.Unpack(stream, &struct1)
//     case 2:
//       packx.Unpack(stream, &struct2)
//     ...
// }
func (packx Packx) Unpack(stream []byte) (Message, error) {
	return UnpackWithMarshaller(stream)
}

// returns the first block's messageID, header, body marshalled stream, error.
func UnPackFromReader(r io.Reader) (int32, []byte, error) {
	buf, e := UnpackToBlockFromReader(r)
	if e != nil {
		return 0, nil, e
	}

	messageID, e := MessageIDOf(buf)
	if e != nil {
		return 0, nil, e
	}

	body, e := BodyBytesOf(buf)
	if e != nil {
		return 0, nil, e
	}
	return messageID, body, nil
}

// a stream from a reader can be apart by protocol.
// FirstBlockOf helps tear apart the first block []byte from reader
func (packx Packx) FirstBlockOf(r io.Reader) ([]byte, error) {
	return FirstBlockOf(r)
}

// Since FirstBlockOf has nothing to do with packx instance, so make it alone,
// for old usage remaining useful, old packx.FirstBlockOf is still useful
func FirstBlockOf(r io.Reader) ([]byte, error) {
	return UnpackToBlockFromReader(r)
}

//// a stream from a buffer which can be apart by protocol.
//// FirstBlockOfBytes helps tear apart the first block []byte from a []byte buffer
//func (packx Packx) FirstBlockOfBytes(buffer []byte) ([]byte, error) {
//	return FirstBlockOfBytes(buffer)
//}
//func FirstBlockOfBytes(buffer []byte) ([]byte, error) {
//	if len(buffer) < 8 {
//		return nil, errors.New(fmt.Sprintf("require buffer length more than 16 but got %d", len(buffer)))
//	}
//	var length = binary.BigEndian.Uint32(buffer[0:4])
//	if len(buffer) < 4+int(length) {
//		return nil, errors.New(fmt.Sprintf("require buffer length more than %d but got %d", 4+int(length), len(buffer)))
//
//	}
//	return buffer[:4+int(length)], nil
//}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func (packx Packx) MessageIDOf(stream []byte) (int32, error) {
	return MessageIDOf(stream)
}

// messageID of a stream.
// Use this to choose which struct for unpacking.
func MessageIDOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	messageID := binary.BigEndian.Uint32(stream[4:8])
	return int32(messageID), nil
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func (packx Packx) LengthOf(stream []byte) (int32, error) {
	return LengthOf(stream)
}

// Length of the stream starting validly.
// Length doesn't include length flag itself, it refers to a valid message length after it.
func LengthOf(stream []byte) (int32, error) {
	if len(stream) < 4 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 4"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return int32(length), nil
}

// Body length of a stream received
func (packx Packx) BodyLengthOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return int32(length) - 8, nil
}

// Body length of a stream received
func BodyLengthOf(stream []byte) (int32, error) {
	if len(stream) < 8 {
		return 0, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	length := binary.BigEndian.Uint32(stream[0:4])
	return int32(length) - 8, nil
}

// body bytes of a block
func (packx Packx) BodyBytesOf(stream []byte) ([]byte, error) {
	return BodyBytesOf(stream)
}

// body bytes of a block
func BodyBytesOf(stream []byte) ([]byte, error) {
	if len(stream) < 8 {
		return []byte{}, errors.New(fmt.Sprintf("stream lenth should be bigger than 8"))
	}
	body := stream[8:len(stream)]
	return body, nil
}

// PackWithMarshaller will encode message into blocks of length,messageID,headerLength,header,bodyLength,body.
// Users don't need to know how pack serializes itself if users use UnpackPWithMarshaller.
//
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [0 0 0 24 0 0 0 1 0 0 0 6 0 0 0 6 2 1 19 18 13 11 11 3 1 23 12 132]
// header: [0 0 0 24]
// mesageID: [0 0 0 1]
// headerLength, bodyLength [0 0 0 6]
// header: [2 1 19 18 13 11]
// body: [11 3 1 23 12 132]
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// []byte -- body                marshal by marshaller
func PackWithMarshaller(message Message) ([]byte, error) {

	var lengthBuf = make([]byte, 4)
	var messageIDBuf = make([]byte, 4)
	binary.BigEndian.PutUint32(messageIDBuf, uint32(message.MessageID))
	var content = make([]byte, 0, 1024)
	content = append(content, messageIDBuf...)
	if message.Body != nil {
		content = append(content, message.Body...)
	}
	binary.BigEndian.PutUint32(lengthBuf, 4+uint32(len(content))) // lengthBuf size 4

	var packet = make([]byte, 0, 1024)
	packet = append(packet, lengthBuf...)
	packet = append(packet, content...)
	return packet, nil
}

//// same as above
//func PackWithMarshallerName(message Message, marshallerName string) ([]byte, error) {
//	var marshaller Marshaller
//	switch marshallerName {
//	case "json":
//		marshaller = JsonMarshaller{}
//	case "xml":
//		marshaller = XmlMarshaller{}
//	case "toml", "tml":
//		marshaller = TomlMarshaller{}
//	case "yaml", "yml":
//		marshaller = YamlMarshaller{}
//	case "protobuf", "proto":
//		marshaller = ProtobufMarshaller{}
//	default:
//		return nil, errors.New("only accept ['json', 'xml', 'toml','yaml','protobuf']")
//	}
//	return PackWithMarshaller(message, marshaller)
//}

// unpack stream from PackWithMarshaller
// If users want to use this protocol across languages, here are the protocol details:
// (they are ordered as list)
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// []byte -- body                marshal by marshaller
func UnpackWithMarshaller(stream []byte) (Message, error) {

	// 包长
	length := binary.BigEndian.Uint32(stream[0:4])
	if length < 8 {
		return Message{}, errors.New("data length is too short!")
	}

	return Message{
		MessageID: int32(binary.BigEndian.Uint32(stream[4:8])),
		Body:      stream[8:length],
	}, nil
}

//// same as above
//func UnpackWithMarshallerName(stream []byte, dest interface{}, marshallerName string) (Message, error) {
//	var marshaller Marshaller
//	switch marshallerName {
//	case "json":
//		marshaller = JsonMarshaller{}
//	case "xml":
//		marshaller = XmlMarshaller{}
//	case "toml", "tml":
//		marshaller = TomlMarshaller{}
//	case "yaml", "yml":
//		marshaller = YamlMarshaller{}
//	case "protobuf", "proto":
//		marshaller = ProtobufMarshaller{}
//	default:
//		return Message{}, errors.New("only accept ['json', 'xml', 'toml','yaml','protobuf']")
//	}
//	return UnpackWithMarshaller(stream, dest, marshaller)
//}

// unpack the first block from the reader.
// protocol is PackWithMarshaller().
// [4]byte -- length             fixed_size,binary big endian encode
// [4]byte -- messageID          fixed_size,binary big endian encode
// [4]byte -- headerLength       fixed_size,binary big endian encode
// [4]byte -- bodyLength         fixed_size,binary big endian encode
// []byte -- header              marshal by json
// []byte -- body                marshal by marshaller
// ussage:
// for {
//     blockBuf, e:= UnpackToBlockFromReader(reader)
// 	   go func(buf []byte){
//         // handle a message block apart
//     }(blockBuf)
//     continue
// }
func UnpackToBlockFromReader(reader io.Reader) ([]byte, error) {
	if reader == nil {
		return nil, errors.New("reader is nil")
	}
	var info = make([]byte, 4, 4)
	if e := readUntil(reader, info); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	length, e := LengthOf(info)
	if e != nil {
		return nil, e
	}
	var content = make([]byte, length, length)
	if e := readUntil(reader, content); e != nil {
		if e == io.EOF {
			return nil, e
		}
		return nil, errorx.Wrap(e)
	}

	return append(info, content...), nil
}

func readUntil(reader io.Reader, buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	var offset int
	for {
		n, e := reader.Read(buf[offset:])
		if e != nil {
			if e == io.EOF {
				return e
			}
			return errorx.Wrap(e)
		}
		offset += n
		if offset >= len(buf) {
			break
		}
	}
	return nil
}

func PackHeartbeat() []byte {
	buf, e := PackWithMarshaller(Message{
		MessageID: DEFAULT_HEARTBEAT_MESSAGEID,
	})
	if e != nil {
		panic(e)
	}
	return buf
}

// pack short signal which only contains messageID
func PackStuff(messageID int32) []byte {
	buf, e := PackWithMarshaller(Message{
		MessageID: messageID,
	})
	if e != nil {
		panic(e)
	}
	return buf
}
