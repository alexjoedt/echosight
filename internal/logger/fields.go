package logger

import (
	"fmt"

	"github.com/google/uuid"
)

type Field struct {
	Type  string
	Key   string
	Value any
}

func Str(k string, v string) Field {
	return Field{
		Type:  "string",
		Key:   k,
		Value: v,
	}
}

func Int(k string, v int) Field {
	return Field{
		Type:  "int",
		Key:   k,
		Value: v,
	}
}

func Any(k string, v any) Field {
	return Field{
		Type:  "any",
		Key:   k,
		Value: v,
	}
}

func UUID(k string, v uuid.UUID) Field {
	return Field{
		Type:  "uuid",
		Key:   k,
		Value: v.String()[:8],
	}
}

func (f Field) Print() string {
	return fmt.Sprintf("%s=%v", f.Key, f.Value)
}
