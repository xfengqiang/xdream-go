package xutil

import (
	"fmt"
	"encoding/json"
)

func JsonStringError(input string, err error) string  {
	if jsonError, ok := err.(*json.SyntaxError); ok {
		line, character, _ := lineAndCharacter(input, int(jsonError.Offset))
		//if lcErr != nil {
		//	fmt.Fprintf(os.Stderr, "Couldn't find the line and character position of the error due to error %v\n", lcErr)
		//}
		return fmt.Sprintf(" Cannot parse JSON schema due to a syntax error at line %d, character %d: %v\n", line, character, jsonError.Error())


	}
	if jsonError, ok := err.(*json.UnmarshalTypeError); ok {
		line, character, _ := lineAndCharacter(input, int(jsonError.Offset))
		//if lcErr != nil {
		//	fmt.Fprintf(os.Stderr, "test %d failed with error: Couldn't find the line and character position of the error due to error %v\n", i+1, lcErr)
		//}
		return fmt.Sprintf("The JSON type '%v' cannot be converted into the Go '%v' type on struct '%s', field '%v'. See input file line %d, character %d\n", jsonError.Value, jsonError.Type.Name(), jsonError.Struct, jsonError.Field, line, character)

	}
	return err.Error()
}

func lineAndCharacter(input string, offset int) (line int, character int, err error) {
	lf := rune(0x0A)

	if offset > len(input) || offset < 0 {
		return 0, 0, fmt.Errorf("Couldn't find offset %d within the input.", offset)
	}

	// Humans tend to count from 1.
	line = 1

	for i, b := range input {
		if b == lf {
			line++
			character = 0
		}
		character++
		if i == offset {
			break
		}
	}

	return line, character, nil
}
