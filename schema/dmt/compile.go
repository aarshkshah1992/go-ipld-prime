package schemadmt

import (
	"github.com/ipld/go-ipld-prime/schema"
)

// This code is broken up into a bunch of individual 'compile' methods,
//  each attached to the type that's their input information.
// However, many of them return distinct concrete types,
//  and so we've just chained it all together with switch statements;
//   creating a separate interface per result type seems just not super relevant.

func (schdmt Schema) Compile() (*schema.TypeSystem, []error) {
	c := &schema.Compiler{}
	c.Init()
	typesdmt := schdmt.FieldTypes()
	for itr := typesdmt.Iterator(); !itr.Done(); {
		tn, t := itr.Next()
		switch t2 := t.AsInterface().(type) {
		case TypeBool:
			c.TypeBool(schema.TypeName(tn.String()))
		case TypeString:
			c.TypeString(schema.TypeName(tn.String()))
		case TypeBytes:
			c.TypeBytes(schema.TypeName(tn.String()))
		case TypeInt:
			c.TypeInt(schema.TypeName(tn.String()))
		case TypeFloat:
			c.TypeFloat(schema.TypeName(tn.String()))
		case TypeLink:
			if t2.FieldExpectedType().Exists() {
				c.TypeLink(schema.TypeName(tn.String()), schema.TypeName(t2.FieldExpectedType().Must().String()))
			} else {
				c.TypeLink(schema.TypeName(tn.String()), "")
			}
		case TypeMap:
			c.TypeMap(
				schema.TypeName(tn.String()),
				schema.TypeName(t2.FieldKeyType().String()),
				t2.FieldValueType().TypeReference(),
				t2.FieldValueNullable().Bool(),
				t2.FieldRepresentation().flip(c),
			)
			// If the field typeReference is TypeDefnInline, that needs a chance to take additional action.
			t2.FieldValueType().flip(c)
		case TypeList:
			c.TypeList(
				schema.TypeName(tn.String()),
				t2.FieldValueType().TypeReference(),
				t2.FieldValueNullable().Bool(),
			)
			// If the field typeReference is TypeDefnInline, that needs a chance to take additional action.
			t2.FieldValueType().flip(c)
		case TypeStruct:
			// Flip fields info from DMT to compiler argument format.
			fields := make([]schema.StructField, t2.FieldFields().Length())
			for itr := t2.FieldFields().Iterator(); !itr.Done(); {
				fname, fdmt := itr.Next()
				fields = append(fields, schema.Compiler{}.MakeStructField(
					schema.StructFieldName(fname.String()),
					fdmt.FieldType().TypeReference(),
					fdmt.FieldOptional().Bool(),
					fdmt.FieldNullable().Bool(),
				))
				// If the field typeReference is TypeDefnInline, that needs a chance to take additional action.
				fdmt.FieldType().flip(c)
			}
			// Flip the representaton strategy DMT to compiler argument format.
			rstrat := func() schema.StructRepresentation {
				switch r := t2.FieldRepresentation().AsInterface().(type) {
				case StructRepresentation_Map:
					return r.flip()
				case StructRepresentation_Tuple:
					return r.flip()
				case StructRepresentation_Stringpairs:
					return r.flip()
				case StructRepresentation_Stringjoin:
					return r.flip()
				case StructRepresentation_Listpairs:
					return r.flip()
				default:
					panic("unreachable")
				}
			}()
			// Feed it all into the compiler.
			c.TypeStruct(
				schema.TypeName(tn.String()),
				schema.Compiler{}.MakeList__StructField(fields...),
				rstrat,
			)
		case TypeUnion:
			// Flip members info from DMT to compiler argument format.
			members := make([]schema.TypeName, t2.FieldMembers().Length())
			for itr := t2.FieldMembers().Iterator(); !itr.Done(); {
				_, memberName := itr.Next()
				members = append(members, schema.TypeName(memberName.String()))
				// n.b. no need to check for TypeDefnInline here, because schemas don't allow those in union defns.
			}
			// Flip the representaton strategy DMT to compiler argument format.
			rstrat := func() schema.UnionRepresentation {
				switch r := t2.FieldRepresentation().AsInterface().(type) {
				case UnionRepresentation_Keyed:
					return r.flip()
				case UnionRepresentation_Kinded:
					return r.flip()
				case UnionRepresentation_Envelope:
					return r.flip()
				case UnionRepresentation_Inline:
					return r.flip()
				case UnionRepresentation_StringPrefix:
					return r.flip()
				case UnionRepresentation_BytePrefix:
					return r.flip()
				default:
					panic("unreachable")
				}
			}()
			// Feed it all into the compiler.
			c.TypeUnion(
				schema.TypeName(tn.String()),
				schema.Compiler{}.MakeList__TypeName(members...),
				rstrat,
			)
		case TypeEnum:
			panic("TODO")
		case TypeCopy:
			panic("no support for 'copy' types.  I might want to reneg on whether these are even part of the schema dmt.")
		default:
			panic("unreachable")
		}
	}
	return c.Compile()
}

// If the typeReference is TypeDefnInline, create the anonymous type and feed it to the compiler.
// It's fine if anonymous type has been seen before; we let dedup of that be handled by the compiler.
func (dmt TypeNameOrInlineDefn) flip(c *schema.Compiler) {
	switch dmt.AsInterface().(type) {
	case TypeDefnInline:
		panic("nyi") // TODO this needs to engage in anonymous type spawning.
	}
}

func (dmt MapRepresentation) flip(c *schema.Compiler) schema.MapRepresentation {
	switch rdmt := dmt.AsInterface().(type) {
	case MapRepresentation_Map:
		return schema.MapRepresentation_Map{}
	case MapRepresentation_Listpairs:
		return schema.MapRepresentation_Listpairs{}
	case MapRepresentation_Stringpairs:
		return c.MakeMapRepresentation_Stringpairs(rdmt.FieldInnerDelim().String(), rdmt.FieldEntryDelim().String())
	default:
		panic("unreachable")
	}
}

func (dmt StructRepresentation_Map) flip() schema.StructRepresentation {
	if !dmt.FieldFields().Exists() {
		return schema.Compiler{}.MakeStructRepresentation_Map(schema.Compiler{}.MakeMap__StructFieldName__StructRepresentation_Map_FieldDetails())
	}
	fields := schema.Compiler{}.StartMap__StructFieldName__StructRepresentation_Map_FieldDetails(int(dmt.FieldFields().Must().Length()))
	for itr := dmt.FieldFields().Must().Iterator(); !itr.Done(); {
		fn, det := itr.Next()
		fields.Append(
			schema.StructFieldName(fn.String()),
			schema.StructRepresentation_Map_FieldDetails{
				Rename: func() string {
					if det.FieldRename().Exists() {
						return det.FieldRename().Must().String()
					}
					return ""
				}(),
				Implicit: nil, // TODO
			},
		)
	}
	return schema.Compiler{}.MakeStructRepresentation_Map(fields.Finish())
}

func (dmt StructRepresentation_Tuple) flip() schema.StructRepresentation {
	panic("TODO")
}

func (dmt StructRepresentation_Stringpairs) flip() schema.StructRepresentation {
	panic("TODO")
}

func (dmt StructRepresentation_Stringjoin) flip() schema.StructRepresentation {
	return schema.Compiler{}.MakeStructRepresentation_Stringjoin(
		dmt.FieldJoin().String(),
		func() (v []schema.StructFieldName) {
			// Maybeism is carried forward here as a nil.
			//  - Precomputing the defaults would require looking at information up-tree, so we leave it to the Compiler to do later.
			//  - The difference between nil (meaning "default") and empty list is still significant; the latter should be able to cause validation errors.
			//  - The carrier types don't have an explicit maybe; a pointer and nil is used as poor maybe.
			// ...

			// OH, MY, GOD.
			// The quickimmut things don't allow you to carry a nil around in a list.
			// The mixture of how sentinels and maybes carry through dmt|carriers|precompile|compile is just insane.
			//
			// And yes: we do need to disambiguate versus zero-len list here: specifying it but being empty: that's an error.
			//  Fwiw: this is the only place in the whole schemaschema where this appears right now.  There are no other optional lists nor optional maps.

			// ... it keeps getting worse.  we can't refer to the type so we can't use this helper function.  lol.
			if !dmt.FieldFieldOrder().Exists() {
				return nil
			}
			for itr := dmt.FieldFieldOrder().Must().Iterator(); !itr.Done(); {
				_, fn := itr.Next()
				v = append(v, schema.StructFieldName(fn.String()))
			}
			return
		}(),
	)
}

func (dmt StructRepresentation_Listpairs) flip() schema.StructRepresentation {
	panic("TODO")
}

func (dmt UnionRepresentation_Keyed) flip() schema.UnionRepresentation {
	ents := schema.Compiler{}.StartMap__String__TypeName(int(dmt.Length()))
	for itr := dmt.Iterator(); !itr.Done(); {
		k, v := itr.Next()
		ents.Append(k.String(), schema.TypeName(v.String()))
	}
	return schema.Compiler{}.MakeUnionRepresentation_Keyed(ents.Finish())
}

func (dmt UnionRepresentation_Kinded) flip() schema.UnionRepresentation {
	panic("TODO")
}

func (dmt UnionRepresentation_Envelope) flip() schema.UnionRepresentation {
	panic("TODO")
}

func (dmt UnionRepresentation_Inline) flip() schema.UnionRepresentation {
	panic("TODO")
}

func (dmt UnionRepresentation_StringPrefix) flip() schema.UnionRepresentation {
	panic("TODO")
}

func (dmt UnionRepresentation_BytePrefix) flip() schema.UnionRepresentation {
	panic("TODO")
}