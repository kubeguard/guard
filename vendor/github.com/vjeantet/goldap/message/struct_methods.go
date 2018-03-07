package message

import (
	"errors"
	"reflect"
)

func (b *Bytes) Bytes() []byte {
	return b.bytes
}

func (l LDAPOID) String() string {
	return string(l)
}

func (l LDAPOID) Bytes() []byte {
	return []byte(l)
}

func (l OCTETSTRING) String() string {
	return string(l)
}

func (l OCTETSTRING) Bytes() []byte {
	return []byte(l)
}

func (l INTEGER) Int() int {
	return int(l)
}
func (l MessageID) Int() int {
	return int(l)
}

func (l ENUMERATED) Int() int {
	return int(l)
}

func (l BOOLEAN) Bool() bool {
	return bool(l)
}

func (l *LDAPMessage) MessageID() MessageID {
	return l.messageID
}
func (l *LDAPMessage) SetMessageID(ID int) {
	l.messageID = MessageID(ID)
}

func (l *LDAPMessage) Controls() *Controls {
	return l.controls
}

func (l *LDAPMessage) ProtocolOp() ProtocolOp {
	return l.protocolOp
}
func (l *LDAPMessage) ProtocolOpName() string {
	return reflect.TypeOf(l.ProtocolOp()).Name()
}
func (l *LDAPMessage) ProtocolOpType() int {
	switch l.protocolOp.(type) {
	case BindRequest:
		return TagBindRequest
	}
	return 0
}

func (b *BindRequest) Name() LDAPDN {
	return b.name
}

func (b *BindRequest) Authentication() AuthenticationChoice {
	return b.authentication
}

func (b *BindRequest) AuthenticationSimple() OCTETSTRING {
	return b.Authentication().(OCTETSTRING)
}

func (b *BindRequest) AuthenticationChoice() string {
	switch b.Authentication().(type) {
	case OCTETSTRING:
		return "simple"
	case SaslCredentials:
		return "sasl"
	}
	return ""
}

func (e *ExtendedRequest) RequestName() LDAPOID {
	return e.requestName
}

func (e *ExtendedRequest) RequestValue() *OCTETSTRING {
	return e.requestValue
}

func (s *SearchRequest) BaseObject() LDAPDN {
	return s.baseObject
}
func (s *SearchRequest) Scope() ENUMERATED {
	return s.scope
}
func (s *SearchRequest) DerefAliases() ENUMERATED {
	return s.derefAliases
}
func (s *SearchRequest) SizeLimit() INTEGER {
	return s.sizeLimit
}

func (s *SearchRequest) TimeLimit() INTEGER {
	return s.timeLimit
}
func (s *SearchRequest) TypesOnly() BOOLEAN {
	return s.typesOnly
}
func (s *SearchRequest) Attributes() AttributeSelection {
	return s.attributes
}

func (s *SearchRequest) Filter() Filter {
	return s.filter
}

func (s *SearchRequest) FilterString() string {
	str, _ := s.decompileFilter(s.Filter())
	return str
}

func (s *SearchRequest) decompileFilter(packet Filter) (ret string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("error decompiling filter")
		}
	}()

	ret = "("
	err = nil
	childStr := ""

	switch f := packet.(type) {
	case FilterAnd:
		ret += "&"
		for _, child := range f {
			childStr, err = s.decompileFilter(child)
			if err != nil {
				return
			}
			ret += childStr
		}
	case FilterOr:
		ret += "|"
		for _, child := range f {
			childStr, err = s.decompileFilter(child)
			if err != nil {
				return
			}
			ret += childStr
		}
	case FilterNot:
		ret += "!"
		childStr, err = s.decompileFilter(f.Filter)
		if err != nil {
			return
		}
		ret += childStr

	case FilterSubstrings:
		ret += string(f.Type_())
		ret += "="
		for _, fs := range f.Substrings() {
			switch fsv := fs.(type) {
			case SubstringInitial:
				ret += string(fsv) + "*"
			case SubstringAny:
				ret += "*" + string(fsv) + "*"
			case SubstringFinal:
				ret += "*" + string(fsv)
			}
		}
	case FilterEqualityMatch:
		ret += string(f.AttributeDesc())
		ret += "="
		ret += string(f.AssertionValue())
	case FilterGreaterOrEqual:
		ret += string(f.AttributeDesc())
		ret += ">="
		ret += string(f.AssertionValue())
	case FilterLessOrEqual:
		ret += string(f.AttributeDesc())
		ret += "<="
		ret += string(f.AssertionValue())
	case FilterPresent:
		// if 0 == len(packet.Children) {
		// 	ret += ber.DecodeString(packet.Data.Bytes())
		// } else {
		// 	ret += ber.DecodeString(packet.Children[0].Data.Bytes())
		// }
		ret += string(f)
		ret += "=*"
	case FilterApproxMatch:
		ret += string(f.AttributeDesc())
		ret += "~="
		ret += string(f.AssertionValue())
	}

	ret += ")"
	return
}

func (c *CompareRequest) Entry() LDAPDN {
	return c.entry
}

func (c *CompareRequest) Ava() *AttributeValueAssertion {
	return &c.ava
}

func (a *AttributeValueAssertion) AttributeDesc() AttributeDescription {
	return a.attributeDesc
}

func (a *AttributeValueAssertion) AssertionValue() AssertionValue {
	return a.assertionValue
}

func (a *AddRequest) Entry() LDAPDN {
	return a.entry
}

func (a *AddRequest) Attributes() AttributeList {
	return a.attributes
}

func (a *Attribute) Type_() AttributeDescription {
	return a.type_
}
func (a *Attribute) Vals() []AttributeValue {
	return a.vals
}

func (m *ModifyRequest) Object() LDAPDN {
	return m.object
}
func (m *ModifyRequest) Changes() []ModifyRequestChange {
	return m.changes
}

func (m *ModifyRequestChange) Operation() ENUMERATED {
	return m.operation
}

func (m *ModifyRequestChange) Modification() *PartialAttribute {
	return &m.modification
}

func (p *PartialAttribute) Type_() AttributeDescription {
	return p.type_
}
func (p *PartialAttribute) Vals() []AttributeValue {
	return p.vals
}

func (c *Control) ControlType() LDAPOID {
	return c.controlType
}

func (c *Control) Criticality() BOOLEAN {
	return c.criticality
}

func (c *Control) ControlValue() *OCTETSTRING {
	return c.controlValue
}

func (a *FilterEqualityMatch) AttributeDesc() AttributeDescription {
	return a.attributeDesc
}

func (a *FilterEqualityMatch) AssertionValue() AssertionValue {
	return a.assertionValue
}

func (a *FilterGreaterOrEqual) AttributeDesc() AttributeDescription {
	return a.attributeDesc
}

func (a *FilterGreaterOrEqual) AssertionValue() AssertionValue {
	return a.assertionValue
}

func (a *FilterLessOrEqual) AttributeDesc() AttributeDescription {
	return a.attributeDesc
}

func (a *FilterLessOrEqual) AssertionValue() AssertionValue {
	return a.assertionValue
}

func (a *FilterApproxMatch) AttributeDesc() AttributeDescription {
	return a.attributeDesc
}

func (a *FilterApproxMatch) AssertionValue() AssertionValue {
	return a.assertionValue
}

func (s *FilterSubstrings) Type_() AttributeDescription {
	return s.type_
}

func (s *FilterSubstrings) Substrings() []Substring {
	return s.substrings
}

func (l *CompareResponse) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *ModifyResponse) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *DelResponse) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *AddResponse) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *SearchResultDone) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *LDAPResult) SetResultCode(code int) {
	l.resultCode = ENUMERATED(code)
}

func (l *LDAPResult) SeMatchedDN(code string) {
	l.matchedDN = LDAPDN(code)
}

func (l *LDAPResult) SetDiagnosticMessage(code string) {
	l.diagnosticMessage = LDAPString(code)
}

func (l *LDAPResult) SetReferral(r *Referral) {
	l.referral = r
}

func (e *ExtendedResponse) SetResponseName(name LDAPOID) {
	e.responseName = &name
}

func NewLDAPMessageWithProtocolOp(po ProtocolOp) *LDAPMessage {
	m := NewLDAPMessage()
	m.protocolOp = po
	return m
}

func (s *SearchResultEntry) SetObjectName(on string) {
	s.objectName = LDAPDN(on)
}

func (s *SearchResultEntry) AddAttribute(name AttributeDescription, values ...AttributeValue) {
	var ea = PartialAttribute{type_: name, vals: values}
	s.attributes.add(ea)
}

func (p *PartialAttributeList) add(a PartialAttribute) {
	*p = append(*p, a)
}
