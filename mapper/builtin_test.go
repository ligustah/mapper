package mapper

//
//func Test_strconvMappers(t *testing.T) {
//	targetType := &GoType{
//		Name: "Target",
//		Pkg:  types.NewPackage("test", "target"),
//	}
//
//	inputType := &GoType{
//		Name: "Input",
//		Pkg:  types.NewPackage("test", "input"),
//	}
//
//	type args struct {
//		generator GeneratorFunc
//		target    *jen.Statement
//		input     *jen.Statement
//	}
//	tests := []struct {
//		Name string
//		args args
//		want string
//	}{
//		{
//			Name: "int struct input",
//			args: args{
//				mapNumericToString("FormatInt", "int64"),
//				jen.Id("out").Dot("Field"),
//				jen.Id("input").Dot("Field"),
//			},
//			want: "out.Field = strconv.FormatInt(int64(input.Field), 10)",
//		},
//		{
//			Name: "int parameter input",
//			args: args{
//				mapNumericToString("FormatInt", "int64"),
//				jen.Id("out").Dot("Field"),
//				jen.Id("input"),
//			},
//			want: "out.Field = strconv.FormatInt(int64(input), 10)",
//		},
//		{
//			Name: "uint struct input",
//			args: args{
//				mapNumericToString("FormatUint", "uint64"),
//				jen.Id("out").Dot("Field"),
//				jen.Id("input"),
//			},
//			want: "out.Field = strconv.FormatUint(uint64(input), 10)",
//		},
//		{
//			Name: "uint struct input",
//			args: args{
//				mapStringToNumeric(nil, "ParseUint", 32),
//				jen.Id("out").Dot("Field"),
//				jen.Id("input"),
//			},
//			want: "if num, err := strconv.ParseUint(input, 10, 32); err != nil {\n\t\treturn test.Target{}, err\n\t} else {\n\t\tout.Field = num\n\t}",
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.Name, func(t *testing.T) {
//			if got := tt.args.generator(targetType, inputType, tt.args.target, tt.args.input).GoString(); !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("mapNumericToString() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
