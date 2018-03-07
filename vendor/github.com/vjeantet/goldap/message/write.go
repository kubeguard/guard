package message

import (
	"fmt"
)

func (b BOOLEAN) write(bytes *Bytes) int {
	return bytes.WritePrimitiveSubBytes(classUniversal, tagBoolean, b)
}

func (b BOOLEAN) writeTagged(bytes *Bytes, class int, tag int) int {
	return bytes.WritePrimitiveSubBytes(class, tag, b)
}

func (i INTEGER) write(bytes *Bytes) int {
	return bytes.WritePrimitiveSubBytes(classUniversal, tagInteger, i)
}

func (i INTEGER) writeTagged(bytes *Bytes, class int, tag int) int {
	return bytes.WritePrimitiveSubBytes(class, tag, i)
}

func (e ENUMERATED) write(bytes *Bytes) int {
	return bytes.WritePrimitiveSubBytes(classUniversal, tagEnum, e)
}

func (e ENUMERATED) writeTagged(bytes *Bytes, class int, tag int) int {
	return bytes.WritePrimitiveSubBytes(class, tag, e)
}

func (o OCTETSTRING) write(bytes *Bytes) int {
	return bytes.WritePrimitiveSubBytes(classUniversal, tagOctetString, o)
}

func (o OCTETSTRING) writeTagged(bytes *Bytes, class int, tag int) int {
	return bytes.WritePrimitiveSubBytes(class, tag, o)
}

//   This appendix is normative.
//
//        Lightweight-Directory-Access-Protocol-V3 {1 3 6 1 1 18}
//        -- Copyright (C) The Internet Society (2006).  This version of
//        -- this ASN.1 module is part of RFC 4511; see the RFC itself
//        -- for full legal notices.
//        DEFINITIONS
//        IMPLICIT TAGS
//        EXTENSIBILITY IMPLIED ::=
//
//        BEGIN
//
//        LDAPMessage ::= SEQUENCE {
//             messageID       MessageID,
//             protocolOp      CHOICE {
//                  bindRequest           BindRequest,
//                  bindResponse          BindResponse,
//                  unbindRequest         UnbindRequest,
//                  searchRequest         SearchRequest,
//                  searchResEntry        SearchResultEntry,
//                  searchResDone         SearchResultDone,
//                  searchResRef          SearchResultReference,
//                  modifyRequest         ModifyRequest,
//                  modifyResponse        ModifyResponse,
//                  addRequest            AddRequest,
//                  addResponse           AddResponse,
//                  delRequest            DelRequest,
//                  delResponse           DelResponse,
//                  modDNRequest          ModifyDNRequest,
//                  modDNResponse         ModifyDNResponse,
//                  compareRequest        CompareRequest,
//                  compareResponse       CompareResponse,
//                  abandonRequest        AbandonRequest,
//                  extendedReq           ExtendedRequest,
//                  extendedResp          ExtendedResponse,
//                  ...,
//                  intermediateResponse  IntermediateResponse },
//             controls       [0] Controls OPTIONAL }
//

func (m *LDAPMessage) Write() (bytes *Bytes, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = LdapError{fmt.Sprintf("Error in LDAPMessage.Write: %s", e)}
		}
	}()
	// Compute the needed size
	totalSize := m.size()
	// Initialize the structure
	bytes = &Bytes{
		bytes:  make([]byte, totalSize),
		offset: totalSize,
	}
	// Go !
	size := 0
	if m.controls != nil {
		size += m.controls.writeTagged(bytes, classContextSpecific, TagLDAPMessageControls)
	}
	size += m.protocolOp.write(bytes)
	size += m.messageID.write(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	// Check
	if size != totalSize || bytes.offset != 0 {
		err = LdapError{fmt.Sprintf("Something went wrong while writing the message ! Size is %d instead of %d, final offset is %d instead of 0", size, totalSize, bytes.offset)}
	}
	return
}

//        MessageID ::= INTEGER (0 ..  maxInt)
//
//        maxInt INTEGER ::= 2147483647 -- (2^^31 - 1) --
//
func (m MessageID) write(bytes *Bytes) int {
	return INTEGER(m).write(bytes)
}
func (m MessageID) writeTagged(bytes *Bytes, class int, tag int) int {
	return INTEGER(m).writeTagged(bytes, class, tag)
}

//        LDAPString ::= OCTET STRING -- UTF-8 encoded,
//                                    -- [ISO10646] characters
func (s LDAPString) write(bytes *Bytes) int {
	return OCTETSTRING(s).write(bytes)
}
func (s LDAPString) writeTagged(bytes *Bytes, class int, tag int) int {
	return OCTETSTRING(s).writeTagged(bytes, class, tag)
}

//
//
//
//
//Sermersheim                 Standards Track                    [Page 54]
//
//
//RFC 4511                         LDAPv3                        June 2006
//
//
//        LDAPOID ::= OCTET STRING -- Constrained to <numericoid>
//                                 -- [RFC4512]
func (l LDAPOID) write(bytes *Bytes) int {
	return OCTETSTRING(l).write(bytes)
}
func (l LDAPOID) writeTagged(bytes *Bytes, class int, tag int) int {
	return OCTETSTRING(l).writeTagged(bytes, class, tag)
}

//
//        LDAPDN ::= LDAPString -- Constrained to <distinguishedName>
//                              -- [RFC4514]
func (l LDAPDN) write(bytes *Bytes) int {
	return LDAPString(l).write(bytes)
}
func (l LDAPDN) writeTagged(bytes *Bytes, class int, tag int) int {
	return LDAPString(l).writeTagged(bytes, class, tag)
}

//
//        RelativeLDAPDN ::= LDAPString -- Constrained to <name-component>
//                                      -- [RFC4514]
func (r RelativeLDAPDN) write(bytes *Bytes) int {
	return LDAPString(r).write(bytes)
}

//
//        AttributeDescription ::= LDAPString
//                                -- Constrained to <attributedescription>
//                                -- [RFC4512]
func (a AttributeDescription) write(bytes *Bytes) int {
	return LDAPString(a).write(bytes)
}
func (a AttributeDescription) writeTagged(bytes *Bytes, class int, tag int) int {
	return LDAPString(a).writeTagged(bytes, class, tag)
}

//
//        AttributeValue ::= OCTET STRING
func (a AttributeValue) write(bytes *Bytes) int {
	return OCTETSTRING(a).write(bytes)
}

//
//        AttributeValueAssertion ::= SEQUENCE {
//             attributeDesc   AttributeDescription,
//             assertionValue  AssertionValue }
func (a AttributeValueAssertion) write(bytes *Bytes) (size int) {
	size += a.assertionValue.write(bytes)
	size += a.attributeDesc.write(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

func (a AttributeValueAssertion) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	size += a.assertionValue.write(bytes)
	size += a.attributeDesc.write(bytes)
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//
//        AssertionValue ::= OCTET STRING
func (a AssertionValue) write(bytes *Bytes) int {
	return OCTETSTRING(a).write(bytes)
}

func (a AssertionValue) writeTagged(bytes *Bytes, class int, tag int) int {
	return OCTETSTRING(a).writeTagged(bytes, class, tag)
}

//
//        PartialAttribute ::= SEQUENCE {
//             type       AttributeDescription,
//             vals       SET OF value AttributeValue }
func (p PartialAttribute) write(bytes *Bytes) (size int) {
	for i := len(p.vals) - 1; i >= 0; i-- {
		size += p.vals[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSet, size)
	size += p.type_.write(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//        Attribute ::= PartialAttribute(WITH COMPONENTS {
//             ...,
//             vals (SIZE(1..MAX))})
func (a Attribute) write(bytes *Bytes) (size int) {
	return PartialAttribute(a).write(bytes)
}

//
//        MatchingRuleId ::= LDAPString
func (m MatchingRuleId) writeTagged(bytes *Bytes, class int, tag int) int {
	return LDAPString(m).writeTagged(bytes, class, tag)
}

//
//        LDAPResult ::= SEQUENCE {
//             resultCode         ENUMERATED {
//                  success                      (0),
//                  operationsError              (1),
//                  protocolError                (2),
//                  timeLimitExceeded            (3),
//                  sizeLimitExceeded            (4),
//                  compareFalse                 (5),
//                  compareTrue                  (6),
//                  authMethodNotSupported       (7),
//                  strongerAuthRequired         (8),
//                       -- 9 reserved --
//                  referral                     (10),
//                  adminLimitExceeded           (11),
//                  unavailableCriticalExtension (12),
//                  confidentialityRequired      (13),
//                  saslBindInProgress           (14),
//
//
//
//Sermersheim                 Standards Track                    [Page 55]
//
//
//RFC 4511                         LDAPv3                        June 2006
//
//
//                  noSuchAttribute              (16),
//                  undefinedAttributeType       (17),
//                  inappropriateMatching        (18),
//                  constraintViolation          (19),
//                  attributeOrValueExists       (20),
//                  invalidAttributeSyntax       (21),
//                       -- 22-31 unused --
//                  noSuchObject                 (32),
//                  aliasProblem                 (33),
//                  invalidDNSyntax              (34),
//                       -- 35 reserved for undefined isLeaf --
//                  aliasDereferencingProblem    (36),
//                       -- 37-47 unused --
//                  inappropriateAuthentication  (48),
//                  invalidCredentials           (49),
//                  insufficientAccessRights     (50),
//                  busy                         (51),
//                  unavailable                  (52),
//                  unwillingToPerform           (53),
//                  loopDetect                   (54),
//                       -- 55-63 unused --
//                  namingViolation              (64),
//                  objectClassViolation         (65),
//                  notAllowedOnNonLeaf          (66),
//                  notAllowedOnRDN              (67),
//                  entryAlreadyExists           (68),
//                  objectClassModsProhibited    (69),
//                       -- 70 reserved for CLDAP --
//                  affectsMultipleDSAs          (71),
//                       -- 72-79 unused --
//                  other                        (80),
//                  ...  },
//             matchedDN          LDAPDN,
//             diagnosticMessage  LDAPString,
//             referral           [3] Referral OPTIONAL }
func (l LDAPResult) write(bytes *Bytes) (size int) {
	size += l.writeComponents(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}
func (l LDAPResult) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	size += l.writeComponents(bytes)
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}
func (l LDAPResult) writeComponents(bytes *Bytes) (size int) {
	if l.referral != nil {
		size += l.referral.writeTagged(bytes, classContextSpecific, TagLDAPResultReferral)
	}
	size += l.diagnosticMessage.write(bytes)
	size += l.matchedDN.write(bytes)
	size += l.resultCode.write(bytes)
	return
}

//
//        Referral ::= SEQUENCE SIZE (1..MAX) OF uri URI
func (r Referral) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	for i := len(r) - 1; i >= 0; i-- {
		size += r[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//
//        URI ::= LDAPString     -- limited to characters permitted in
//                               -- URIs
func (u URI) write(bytes *Bytes) int {
	return LDAPString(u).write(bytes)
}

//
//        Controls ::= SEQUENCE OF control Control
func (c Controls) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	for i := len(c) - 1; i >= 0; i-- {
		size += c[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//
//        Control ::= SEQUENCE {
//             controlType             LDAPOID,
//             criticality             BOOLEAN DEFAULT FALSE,
//             controlValue            OCTET STRING OPTIONAL }
func (c Control) write(bytes *Bytes) (size int) {
	if c.controlValue != nil {
		size += c.controlValue.write(bytes)
	}
	if c.criticality != BOOLEAN(false) {
		size += c.criticality.write(bytes)
	}
	size += c.controlType.write(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//
//
//
//Sermersheim                 Standards Track                    [Page 56]
//
//
//RFC 4511                         LDAPv3                        June 2006
//
//
//        BindRequest ::= [APPLICATION 0] SEQUENCE {
//             version                 INTEGER (1 ..  127),
//             name                    LDAPDN,
//             authentication          AuthenticationChoice }
func (b BindRequest) write(bytes *Bytes) (size int) {
	switch b.authentication.(type) {
	case OCTETSTRING:
		size += b.authentication.(OCTETSTRING).writeTagged(bytes, classContextSpecific, TagAuthenticationChoiceSimple)
	case SaslCredentials:
		size += b.authentication.(SaslCredentials).writeTagged(bytes, classContextSpecific, TagAuthenticationChoiceSaslCredentials)
	default:
		panic(fmt.Sprintf("Unknown authentication choice: %#v", b.authentication))
	}
	size += b.name.write(bytes)
	size += b.version.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagBindRequest, size)
	return
}

//
//        AuthenticationChoice ::= CHOICE {
//             simple                  [0] OCTET STRING,
//                                     -- 1 and 2 reserved
//             sasl                    [3] SaslCredentials,
//             ...  }

//
//        SaslCredentials ::= SEQUENCE {
//             mechanism               LDAPString,
//             credentials             OCTET STRING OPTIONAL }
//
func (s SaslCredentials) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	if s.credentials != nil {
		size += s.credentials.write(bytes)
	}
	size += s.mechanism.write(bytes)
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//        BindResponse ::= [APPLICATION 1] SEQUENCE {
//             COMPONENTS OF LDAPResult,
//             serverSaslCreds    [7] OCTET STRING OPTIONAL }
func (b BindResponse) write(bytes *Bytes) (size int) {
	if b.serverSaslCreds != nil {
		size += b.serverSaslCreds.writeTagged(bytes, classContextSpecific, TagBindResponseServerSaslCreds)
	}
	size += b.LDAPResult.writeComponents(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagBindResponse, size)
	return
}

//
//        UnbindRequest ::= [APPLICATION 2] NULL
func (u UnbindRequest) write(bytes *Bytes) (size int) {
	size += bytes.WriteTagAndLength(classApplication, isNotCompound, TagUnbindRequest, 0)
	return
}

//
//        SearchRequest ::= [APPLICATION 3] SEQUENCE {
//             baseObject      LDAPDN,
//             scope           ENUMERATED {
//                  baseObject              (0),
//                  singleLevel             (1),
//                  wholeSubtree            (2),
//                  ...  },
//             derefAliases    ENUMERATED {
//                  neverDerefAliases       (0),
//                  derefInSearching        (1),
//                  derefFindingBaseObj     (2),
//                  derefAlways             (3) },
//             sizeLimit       INTEGER (0 ..  maxInt),
//             timeLimit       INTEGER (0 ..  maxInt),
//             typesOnly       BOOLEAN,
//             filter          Filter,
//             attributes      AttributeSelection }
func (s SearchRequest) write(bytes *Bytes) (size int) {
	size += s.attributes.write(bytes)
	size += s.filter.write(bytes)
	size += s.typesOnly.write(bytes)
	size += s.timeLimit.write(bytes)
	size += s.sizeLimit.write(bytes)
	size += s.derefAliases.write(bytes)
	size += s.scope.write(bytes)
	size += s.baseObject.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagSearchRequest, size)
	return
}

//
//        AttributeSelection ::= SEQUENCE OF selector LDAPString
//                       -- The LDAPString is constrained to
//                       -- <attributeSelector> in Section 4.5.1.8
func (a AttributeSelection) write(bytes *Bytes) (size int) {
	for i := len(a) - 1; i >= 0; i-- {
		size += a[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//        Filter ::= CHOICE {
//             and             [0] SET SIZE (1..MAX) OF filter Filter,
//             or              [1] SET SIZE (1..MAX) OF filter Filter,
//             not             [2] Filter,
//             equalityMatch   [3] AttributeValueAssertion,
//
//
//
//Sermersheim                 Standards Track                    [Page 57]
//
//
//RFC 4511                         LDAPv3                        June 2006
//
//
//             substrings      [4] SubstringFilter,
//             greaterOrEqual  [5] AttributeValueAssertion,
//             lessOrEqual     [6] AttributeValueAssertion,
//             present         [7] AttributeDescription,
//             approxMatch     [8] AttributeValueAssertion,
//             extensibleMatch [9] MatchingRuleAssertion,
//             ...  }

//             and             [0] SET SIZE (1..MAX) OF filter Filter,
func (f FilterAnd) write(bytes *Bytes) (size int) {
	for i := len(f) - 1; i >= 0; i-- {
		size += f[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classContextSpecific, isCompound, TagFilterAnd, size)
	return
}

//             or              [1] SET SIZE (1..MAX) OF filter Filter,
func (f FilterOr) write(bytes *Bytes) (size int) {
	for i := len(f) - 1; i >= 0; i-- {
		size += f[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classContextSpecific, isCompound, TagFilterOr, size)
	return
}

//             not             [2] Filter,
func (f FilterNot) write(bytes *Bytes) (size int) {
	size = f.Filter.write(bytes)
	size += bytes.WriteTagAndLength(classContextSpecific, isCompound, TagFilterNot, size)
	return
}

//             equalityMatch   [3] AttributeValueAssertion,
func (f FilterEqualityMatch) write(bytes *Bytes) int {
	return AttributeValueAssertion(f).writeTagged(bytes, classContextSpecific, TagFilterEqualityMatch)
}

//             substrings      [4] SubstringFilter,
func (f FilterSubstrings) write(bytes *Bytes) int {
	return SubstringFilter(f).writeTagged(bytes, classContextSpecific, TagFilterSubstrings)
}

//             greaterOrEqual  [5] AttributeValueAssertion,
func (f FilterGreaterOrEqual) write(bytes *Bytes) int {
	return AttributeValueAssertion(f).writeTagged(bytes, classContextSpecific, TagFilterGreaterOrEqual)
}

//             lessOrEqual     [6] AttributeValueAssertion,
func (f FilterLessOrEqual) write(bytes *Bytes) int {
	return AttributeValueAssertion(f).writeTagged(bytes, classContextSpecific, TagFilterLessOrEqual)
}

//             present         [7] AttributeDescription,
func (f FilterPresent) write(bytes *Bytes) int {
	return AttributeDescription(f).writeTagged(bytes, classContextSpecific, TagFilterPresent)
}

//             approxMatch     [8] AttributeValueAssertion,
func (f FilterApproxMatch) write(bytes *Bytes) int {
	return AttributeValueAssertion(f).writeTagged(bytes, classContextSpecific, TagFilterApproxMatch)
}

//             extensibleMatch [9] MatchingRuleAssertion,
func (f FilterExtensibleMatch) write(bytes *Bytes) int {
	return MatchingRuleAssertion(f).writeTagged(bytes, classContextSpecific, TagFilterExtensibleMatch)
}

//
//        SubstringFilter ::= SEQUENCE {
//             type           AttributeDescription,
//             substrings     SEQUENCE SIZE (1..MAX) OF substring CHOICE {
//                  initial [0] AssertionValue,  -- can occur at most once
//                  any     [1] AssertionValue,
//                  final   [2] AssertionValue } -- can occur at most once
//             }
func (s SubstringFilter) write(bytes *Bytes) (size int) {
	return s.writeTagged(bytes, classUniversal, tagSequence)
}
func (s SubstringFilter) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	for i := len(s.substrings) - 1; i >= 0; i-- {
		substring := s.substrings[i]
		switch substring.(type) {
		case SubstringInitial:
			size += AssertionValue(substring.(SubstringInitial)).writeTagged(bytes, classContextSpecific, TagSubstringInitial)
		case SubstringAny:
			size += AssertionValue(substring.(SubstringAny)).writeTagged(bytes, classContextSpecific, TagSubstringAny)
		case SubstringFinal:
			size += AssertionValue(substring.(SubstringFinal)).writeTagged(bytes, classContextSpecific, TagSubstringFinal)
		default:
			panic("Unknown type for SubstringFilter substring")
		}
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	size += s.type_.write(bytes)
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//
//        MatchingRuleAssertion ::= SEQUENCE {
//             matchingRule    [1] MatchingRuleId OPTIONAL,
//             type            [2] AttributeDescription OPTIONAL,
//             matchValue      [3] AssertionValue,
//             dnAttributes    [4] BOOLEAN DEFAULT FALSE }
func (m MatchingRuleAssertion) write(bytes *Bytes) (size int) {
	return m.writeTagged(bytes, classUniversal, tagSequence)
}
func (m MatchingRuleAssertion) writeTagged(bytes *Bytes, class int, tag int) (size int) {
	if m.dnAttributes != BOOLEAN(false) {
		size += m.dnAttributes.writeTagged(bytes, classContextSpecific, TagMatchingRuleAssertionDnAttributes)
	}
	size += m.matchValue.writeTagged(bytes, classContextSpecific, TagMatchingRuleAssertionMatchValue)
	if m.type_ != nil {
		size += m.type_.writeTagged(bytes, classContextSpecific, TagMatchingRuleAssertionType)
	}
	if m.matchingRule != nil {
		size += m.matchingRule.writeTagged(bytes, classContextSpecific, TagMatchingRuleAssertionMatchingRule)
	}
	size += bytes.WriteTagAndLength(class, isCompound, tag, size)
	return
}

//
//        SearchResultEntry ::= [APPLICATION 4] SEQUENCE {
//             objectName      LDAPDN,
//             attributes      PartialAttributeList }
func (s SearchResultEntry) write(bytes *Bytes) (size int) {
	size += s.attributes.write(bytes)
	size += s.objectName.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagSearchResultEntry, size)
	return
}

//
//        PartialAttributeList ::= SEQUENCE OF
//                             partialAttribute PartialAttribute
func (p PartialAttributeList) write(bytes *Bytes) (size int) {
	for i := len(p) - 1; i >= 0; i-- {
		size += p[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//        SearchResultReference ::= [APPLICATION 19] SEQUENCE
//                                  SIZE (1..MAX) OF uri URI
func (s SearchResultReference) write(bytes *Bytes) (size int) {
	for i := len(s) - 1; i >= 0; i-- {
		size += s[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagSearchResultReference, size)
	return
}

//
//        SearchResultDone ::= [APPLICATION 5] LDAPResult
func (s SearchResultDone) write(bytes *Bytes) int {
	return LDAPResult(s).writeTagged(bytes, classApplication, TagSearchResultDone)
}

//
//        ModifyRequest ::= [APPLICATION 6] SEQUENCE {
//             object          LDAPDN,
//             changes         SEQUENCE OF change SEQUENCE {
//                  operation       ENUMERATED {
//                       add     (0),
//                       delete  (1),
//                       replace (2),
//                       ...  },
//                  modification    PartialAttribute } }
func (m ModifyRequest) write(bytes *Bytes) (size int) {
	for i := len(m.changes) - 1; i >= 0; i-- {
		size += m.changes[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	size += m.object.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagModifyRequest, size)
	return
}

func (m ModifyRequestChange) write(bytes *Bytes) (size int) {
	size += m.modification.write(bytes)
	size += m.operation.write(bytes)
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//        ModifyResponse ::= [APPLICATION 7] LDAPResult
func (m ModifyResponse) write(bytes *Bytes) int {
	return LDAPResult(m).writeTagged(bytes, classApplication, TagModifyResponse)
}

//
//
//
//
//
//
//Sermersheim                 Standards Track                    [Page 58]
//
//
//RFC 4511                         LDAPv3                        June 2006
//
//
//        AddRequest ::= [APPLICATION 8] SEQUENCE {
//             entry           LDAPDN,
//             attributes      AttributeList }
func (a AddRequest) write(bytes *Bytes) (size int) {
	size += a.attributes.write(bytes)
	size += a.entry.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagAddRequest, size)
	return
}

//
//        AttributeList ::= SEQUENCE OF attribute Attribute
func (a AttributeList) write(bytes *Bytes) (size int) {
	for i := len(a) - 1; i >= 0; i-- {
		size += a[i].write(bytes)
	}
	size += bytes.WriteTagAndLength(classUniversal, isCompound, tagSequence, size)
	return
}

//
//        AddResponse ::= [APPLICATION 9] LDAPResult
func (a AddResponse) write(bytes *Bytes) int {
	return LDAPResult(a).writeTagged(bytes, classApplication, TagAddResponse)
}

//
//        DelRequest ::= [APPLICATION 10] LDAPDN
func (d DelRequest) write(bytes *Bytes) int {
	return LDAPDN(d).writeTagged(bytes, classApplication, TagDelRequest)
}

//
//        DelResponse ::= [APPLICATION 11] LDAPResult
func (d DelResponse) write(bytes *Bytes) int {
	return LDAPResult(d).writeTagged(bytes, classApplication, TagDelResponse)
}

//
//        ModifyDNRequest ::= [APPLICATION 12] SEQUENCE {
//             entry           LDAPDN,
//             newrdn          RelativeLDAPDN,
//             deleteoldrdn    BOOLEAN,
//             newSuperior     [0] LDAPDN OPTIONAL }
func (m ModifyDNRequest) write(bytes *Bytes) (size int) {
	if m.newSuperior != nil {
		size += m.newSuperior.writeTagged(bytes, classContextSpecific, TagModifyDNRequestNewSuperior)
	}
	size += m.deleteoldrdn.write(bytes)
	size += m.newrdn.write(bytes)
	size += m.entry.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagModifyDNRequest, size)
	return
}

//
//        ModifyDNResponse ::= [APPLICATION 13] LDAPResult
func (m ModifyDNResponse) write(bytes *Bytes) int {
	return LDAPResult(m).writeTagged(bytes, classApplication, TagModifyDNResponse)
}

//
//        CompareRequest ::= [APPLICATION 14] SEQUENCE {
//             entry           LDAPDN,
//             ava             AttributeValueAssertion }
func (c CompareRequest) write(bytes *Bytes) (size int) {
	size += c.ava.write(bytes)
	size += c.entry.write(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagCompareRequest, size)
	return
}

//
//        CompareResponse ::= [APPLICATION 15] LDAPResult
func (c CompareResponse) write(bytes *Bytes) int {
	return LDAPResult(c).writeTagged(bytes, classApplication, TagCompareResponse)
}

//
//        AbandonRequest ::= [APPLICATION 16] MessageID
func (a AbandonRequest) write(bytes *Bytes) int {
	return MessageID(a).writeTagged(bytes, classApplication, TagAbandonRequest)
}

//
//        ExtendedRequest ::= [APPLICATION 23] SEQUENCE {
//             requestName      [0] LDAPOID,
//             requestValue     [1] OCTET STRING OPTIONAL }
func (e ExtendedRequest) write(bytes *Bytes) (size int) {
	if e.requestValue != nil {
		size += e.requestValue.writeTagged(bytes, classContextSpecific, TagExtendedRequestValue)
	}
	size += e.requestName.writeTagged(bytes, classContextSpecific, TagExtendedRequestName)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagExtendedRequest, size)
	return
}

//
//        ExtendedResponse ::= [APPLICATION 24] SEQUENCE {
//             COMPONENTS OF LDAPResult,
//             responseName     [10] LDAPOID OPTIONAL,
//             responseValue    [11] OCTET STRING OPTIONAL }
func (e ExtendedResponse) write(bytes *Bytes) (size int) {
	if e.responseValue != nil {
		size += e.responseValue.writeTagged(bytes, classContextSpecific, TagExtendedResponseValue)
	}
	if e.responseName != nil {
		size += e.responseName.writeTagged(bytes, classContextSpecific, TagExtendedResponseName)
	}
	size += e.LDAPResult.writeComponents(bytes)
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagExtendedResponse, size)
	return
}

//
//        IntermediateResponse ::= [APPLICATION 25] SEQUENCE {
//             responseName     [0] LDAPOID OPTIONAL,
//             responseValue    [1] OCTET STRING OPTIONAL }
func (i IntermediateResponse) write(bytes *Bytes) (size int) {
	if i.responseValue != nil {
		size += i.responseValue.writeTagged(bytes, classContextSpecific, TagIntermediateResponseValue)
	}
	if i.responseName != nil {
		size += i.responseName.writeTagged(bytes, classContextSpecific, TagIntermediateResponseName)
	}
	size += bytes.WriteTagAndLength(classApplication, isCompound, TagIntermediateResponse, size)
	return
}

//
//        END
//
