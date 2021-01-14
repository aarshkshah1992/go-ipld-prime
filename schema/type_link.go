package schema

import (
	"github.com/ipld/go-ipld-prime"
)

type TypeLink struct {
	ts              *TypeSystem
	name            TypeName
	expectedTypeRef TypeName // can be empty
}

// -- Type interface satisfaction -->

var _ Type = (*TypeLink)(nil)

func (t *TypeLink) TypeSystem() *TypeSystem {
	return t.ts
}

func (TypeLink) TypeKind() TypeKind {
	return TypeKind_Link
}

func (t *TypeLink) Name() TypeName {
	return t.name
}

func (t TypeLink) RepresentationBehavior() ipld.Kind {
	return ipld.Kind_Link
}

// -- specific to TypeLink -->

// HasExpectedType returns true if the link has a hint about the type it references.
func (t *TypeLink) HasExpectedType() bool {
	return t.expectedTypeRef != ""
}

// ExpectedType returns the type which is expected for the node on the other side of the link.
// Nil is returned if there is no information about the expected type
// (which may be interpreted as "any").
func (t *TypeLink) ExpectedType() Type {
	if !t.HasExpectedType() {
		return nil
	}
	return t.ts.types[TypeReference(t.expectedTypeRef)]
}