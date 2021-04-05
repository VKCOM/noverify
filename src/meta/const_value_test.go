package meta

import "testing"

func TestConstantValueDecodeEncode(t *testing.T) {
	testCases := []ConstValue{
		{Type: String, Value: "world"},
		{Type: Integer, Value: int64(5)},
		{Type: Float, Value: 5.56},
		{Type: String, Value: "hello"},
		{Type: Float, Value: 124.67},
		{Type: Integer, Value: int64(50000000)},
	}

	for _, testCase := range testCases {
		// encode this
		encoded, err := testCase.GobEncode()
		if err != nil {
			t.Errorf("unexpected error \"%s\"", err)
		}

		// decode this
		decoded := ConstValue{}
		err = decoded.GobDecode(encoded)
		if err != nil {
			t.Errorf("unexpected error \"%s\"", err)
		}

		// compare
		if decoded.Type != testCase.Type || decoded.Value != testCase.Value {
			t.Error("error decode, objects not equal")
		}
	}
}
