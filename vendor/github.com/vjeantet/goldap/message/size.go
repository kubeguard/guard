package message

import (
	"fmt"
)

func (b BOOLEAN) size() int {
	return SizePrimitiveSubBytes(tagBoolean, b)
}

func (b BOOLEAN) sizeTagged(tag int) int {
	return SizePrimitiveSubBytes(tag, b)
}

func (i INTEGER) size() int {
	return SizePrimitiveSubBytes(tagInteger, i)
}

func (i INTEGER) sizeTagged(tag int) int {
	return SizePrimitiveSubBytes(tag, i)
}

func (e ENUMERATED) size() int {
	return SizePrimitiveSubBytes(tagEnum, e)
}

func (o OCTETSTRING) size() int {
	return SizePrimitiveSubBytes(tagOctetString, o)
}

func (o OCTETSTRING) sizeTagged(tag int) int {
	return SizePrimitiveSubBytes(tag, o)
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

func (m *LDAPMessage) size() (size int) {
	size += m.messageID.size()
	size += m.protocolOp.size()
	if m.controls != nil {
		size += m.controls.sizeTagged(TagLDAPMessageControls)
	}
	size += sizeTagAndLength(tagSequence, size)
	return
}

//        MessageID ::= INTEGER (0 ..  maxInt)
//
//        maxInt INTEGER ::= 2147483647 -- (2^^31 - 1) --
//
func (m MessageID) size() int {
	return INTEGER(m).size()
}
func (m MessageID) sizeTagged(tag int) int {
	return INTEGER(m).sizeTagged(tag)
}

//        LDAPString ::= OCTET STRING -- UTF-8 encoded,
//                                    -- [ISO10646] characters
func (s LDAPString) size() int {
	return OCTETSTRING(s).size()
}
func (s LDAPString) sizeTagged(tag int) int {
	return OCTETSTRING(s).sizeTagged(tag)
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
func (l LDAPOID) size() int {
	return OCTETSTRING(l).size()
}
func (l LDAPOID) sizeTagged(tag int) int {
	return OCTETSTRING(l).sizeTagged(tag)
}

//
//        LDAPDN ::= LDAPString -- Constrained to <distinguishedName>
//                              -- [RFC4514]
func (l LDAPDN) size() int {
	return LDAPString(l).size()
}
func (l LDAPDN) sizeTagged(tag int) int {
	return LDAPString(l).sizeTagged(tag)
}

//
//        RelativeLDAPDN ::= LDAPString -- Constrained to <name-component>
//                                      -- [RFC4514]
func (r RelativeLDAPDN) size() int {
	return LDAPString(r).size()
}

//
//        AttributeDescription ::= LDAPString
//                                -- Constrained to <attributedescription>
//                                -- [RFC4512]
func (a AttributeDescription) size() int {
	return LDAPString(a).size()
}
func (a AttributeDescription) sizeTagged(tag int) int {
	return LDAPString(a).sizeTagged(tag)
}

//
//        AttributeValue ::= OCTET STRING
func (a AttributeValue) size() int {
	return OCTETSTRING(a).size()
}

//
//        AttributeValueAssertion ::= SEQUENCE {
//             attributeDesc   AttributeDescription,
//             assertionValue  AssertionValue }
func (a AttributeValueAssertion) size() (size int) {
	size += a.attributeDesc.size()
	size += a.assertionValue.size()
	size += sizeTagAndLength(tagSequence, size)
	return
}

func (a AttributeValueAssertion) sizeTagged(tag int) (size int) {
	size += a.attributeDesc.size()
	size += a.assertionValue.size()
	size += sizeTagAndLength(tag, size)
	return
}

//
//        AssertionValue ::= OCTET STRING
func (a AssertionValue) size() int {
	return OCTETSTRING(a).size()
}

func (a AssertionValue) sizeTagged(tag int) int {
	return OCTETSTRING(a).sizeTagged(tag)
}

//
//        PartialAttribute ::= SEQUENCE {
//             type       AttributeDescription,
//             vals       SET OF value AttributeValue }
func (p PartialAttribute) size() (size int) {
	for _, value := range p.vals {
		size += value.size()
	}
	size += sizeTagAndLength(tagSet, size)
	size += p.type_.size()
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        Attribute ::= PartialAttribute(WITH COMPONENTS {
//             ...,
//             vals (SIZE(1..MAX))})
func (a Attribute) size() (size int) {
	return PartialAttribute(a).size()
}

//
//        MatchingRuleId ::= LDAPString
func (m MatchingRuleId) sizeTagged(tag int) int {
	return LDAPString(m).sizeTagged(tag)
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
func (l LDAPResult) size() (size int) {
	size += l.sizeComponents()
	size += sizeTagAndLength(tagSequence, size)
	return
}
func (l LDAPResult) sizeTagged(tag int) (size int) {
	size += l.sizeComponents()
	size += sizeTagAndLength(tag, size)
	return
}
func (l LDAPResult) sizeComponents() (size int) {
	if l.referral != nil {
		size += l.referral.sizeTagged(TagLDAPResultReferral)
	}
	size += l.diagnosticMessage.size()
	size += l.matchedDN.size()
	size += l.resultCode.size()
	return
}

//
//        Referral ::= SEQUENCE SIZE (1..MAX) OF uri URI
func (r Referral) sizeTagged(tag int) (size int) {
	for _, uri := range r {
		size += uri.size()
	}
	size += sizeTagAndLength(tag, size)
	return
}

//
//        URI ::= LDAPString     -- limited to characters permitted in
//                               -- URIs
func (u URI) size() int {
	return LDAPString(u).size()
}

//
//        Controls ::= SEQUENCE OF control Control
func (c Controls) sizeTagged(tag int) (size int) {
	for _, control := range c {
		size += control.size()
	}
	size += sizeTagAndLength(tag, size)
	return
}

//
//        Control ::= SEQUENCE {
//             controlType             LDAPOID,
//             criticality             BOOLEAN DEFAULT FALSE,
//             controlValue            OCTET STRING OPTIONAL }
func (c Control) size() (size int) {
	if c.controlValue != nil {
		size += c.controlValue.size()
	}
	if c.criticality != BOOLEAN(false) {
		size += c.criticality.size()
	}
	size += c.controlType.size()
	size += sizeTagAndLength(tagSequence, size)
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
func (b BindRequest) size() (size int) {
	size += b.version.size()
	size += b.name.size()
	switch b.authentication.(type) {
	case OCTETSTRING:
		size += b.authentication.(OCTETSTRING).sizeTagged(TagAuthenticationChoiceSimple)
	case SaslCredentials:
		size += b.authentication.(SaslCredentials).sizeTagged(TagAuthenticationChoiceSaslCredentials)
	default:
		panic(fmt.Sprintf("Unknown authentication choice: %#v", b.authentication))
	}

	size += sizeTagAndLength(TagBindRequest, size)
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
func (s SaslCredentials) sizeTagged(tag int) (size int) {
	if s.credentials != nil {
		size += s.credentials.size()
	}
	size += s.mechanism.size()
	size += sizeTagAndLength(tag, size)
	return
}

//        BindResponse ::= [APPLICATION 1] SEQUENCE {
//             COMPONENTS OF LDAPResult,
//             serverSaslCreds    [7] OCTET STRING OPTIONAL }
func (b BindResponse) size() (size int) {
	if b.serverSaslCreds != nil {
		size += b.serverSaslCreds.sizeTagged(TagBindResponseServerSaslCreds)
	}
	size += b.LDAPResult.sizeComponents()
	size += sizeTagAndLength(TagBindResponse, size)
	return
}

//
//        UnbindRequest ::= [APPLICATION 2] NULL
func (u UnbindRequest) size() (size int) {
	size = sizeTagAndLength(TagUnbindRequest, 0)
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
func (s SearchRequest) size() (size int) {
	size += s.baseObject.size()
	size += s.scope.size()
	size += s.derefAliases.size()
	size += s.sizeLimit.size()
	size += s.timeLimit.size()
	size += s.typesOnly.size()
	size += s.filter.size()
	size += s.attributes.size()
	size += sizeTagAndLength(TagSearchRequest, size)
	return
}

//
//        AttributeSelection ::= SEQUENCE OF selector LDAPString
//                       -- The LDAPString is constrained to
//                       -- <attributeSelector> in Section 4.5.1.8
func (a AttributeSelection) size() (size int) {
	for _, selector := range a {
		size += selector.size()
	}
	size += sizeTagAndLength(tagSequence, size)
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
func (f FilterAnd) size() (size int) {
	for _, filter := range f {
		size += filter.size()
	}
	size += sizeTagAndLength(TagFilterAnd, size)
	return
}

//             or              [1] SET SIZE (1..MAX) OF filter Filter,
func (f FilterOr) size() (size int) {
	for _, filter := range f {
		size += filter.size()
	}
	size += sizeTagAndLength(TagFilterOr, size)
	return
}

//             not             [2] Filter,
func (f FilterNot) size() (size int) {
	size = f.Filter.size()
	size += sizeTagAndLength(tagSequence, size)
	return
}

//             equalityMatch   [3] AttributeValueAssertion,
func (f FilterEqualityMatch) size() int {
	return AttributeValueAssertion(f).sizeTagged(TagFilterEqualityMatch)
}

//             substrings      [4] SubstringFilter,
func (f FilterSubstrings) size() int {
	return SubstringFilter(f).sizeTagged(TagFilterSubstrings)
}

//             greaterOrEqual  [5] AttributeValueAssertion,
func (f FilterGreaterOrEqual) size() int {
	return AttributeValueAssertion(f).sizeTagged(TagFilterGreaterOrEqual)
}

//             lessOrEqual     [6] AttributeValueAssertion,
func (f FilterLessOrEqual) size() int {
	return AttributeValueAssertion(f).sizeTagged(TagFilterLessOrEqual)
}

//             present         [7] AttributeDescription,
func (f FilterPresent) size() int {
	return AttributeDescription(f).sizeTagged(TagFilterPresent)
}

//             approxMatch     [8] AttributeValueAssertion,
func (f FilterApproxMatch) size() int {
	return AttributeValueAssertion(f).sizeTagged(TagFilterApproxMatch)
}

//             extensibleMatch [9] MatchingRuleAssertion,
func (f FilterExtensibleMatch) size() int {
	return MatchingRuleAssertion(f).sizeTagged(TagFilterExtensibleMatch)
}

//
//        SubstringFilter ::= SEQUENCE {
//             type           AttributeDescription,
//             substrings     SEQUENCE SIZE (1..MAX) OF substring CHOICE {
//                  initial [0] AssertionValue,  -- can occur at most once
//                  any     [1] AssertionValue,
//                  final   [2] AssertionValue } -- can occur at most once
//             }
func (s SubstringFilter) size() (size int) {
	return s.sizeTagged(tagSequence)
}
func (s SubstringFilter) sizeTagged(tag int) (size int) {
	for _, substring := range s.substrings {
		switch substring.(type) {
		case SubstringInitial:
			size += AssertionValue(substring.(SubstringInitial)).sizeTagged(TagSubstringInitial)
		case SubstringAny:
			size += AssertionValue(substring.(SubstringAny)).sizeTagged(TagSubstringAny)
		case SubstringFinal:
			size += AssertionValue(substring.(SubstringFinal)).sizeTagged(TagSubstringFinal)
		default:
			panic("Unknown type for SubstringFilter substring")
		}
	}
	size += sizeTagAndLength(tagSequence, size)
	size += s.type_.size()
	size += sizeTagAndLength(tag, size)
	return
}

//
//        MatchingRuleAssertion ::= SEQUENCE {
//             matchingRule    [1] MatchingRuleId OPTIONAL,
//             type            [2] AttributeDescription OPTIONAL,
//             matchValue      [3] AssertionValue,
//             dnAttributes    [4] BOOLEAN DEFAULT FALSE }
func (m MatchingRuleAssertion) size() (size int) {
	return m.sizeTagged(tagSequence)
}
func (m MatchingRuleAssertion) sizeTagged(tag int) (size int) {
	if m.matchingRule != nil {
		size += m.matchingRule.sizeTagged(TagMatchingRuleAssertionMatchingRule)
	}
	if m.type_ != nil {
		size += m.type_.sizeTagged(TagMatchingRuleAssertionType)
	}
	size += m.matchValue.sizeTagged(TagMatchingRuleAssertionMatchValue)
	if m.dnAttributes != BOOLEAN(false) {
		size += m.dnAttributes.sizeTagged(TagMatchingRuleAssertionDnAttributes)
	}
	size += sizeTagAndLength(tag, size)
	return
}

//
//        SearchResultEntry ::= [APPLICATION 4] SEQUENCE {
//             objectName      LDAPDN,
//             attributes      PartialAttributeList }
func (s SearchResultEntry) size() (size int) {
	size += s.objectName.size()
	size += s.attributes.size()
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        PartialAttributeList ::= SEQUENCE OF
//                             partialAttribute PartialAttribute
func (p PartialAttributeList) size() (size int) {
	for _, att := range p {
		size += att.size()
	}
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        SearchResultReference ::= [APPLICATION 19] SEQUENCE
//                                  SIZE (1..MAX) OF uri URI
func (s SearchResultReference) size() (size int) {
	for _, uri := range s {
		size += uri.size()
	}
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        SearchResultDone ::= [APPLICATION 5] LDAPResult
func (s SearchResultDone) size() int {
	return LDAPResult(s).sizeTagged(TagSearchResultDone)
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
func (m ModifyRequest) size() (size int) {
	for _, change := range m.changes {
		size += change.size()
	}
	size += sizeTagAndLength(tagSequence, size)
	size += m.object.size()
	size += sizeTagAndLength(TagModifyRequest, size)
	return
}

func (m ModifyRequestChange) size() (size int) {
	size += m.operation.size()
	size += m.modification.size()
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        ModifyResponse ::= [APPLICATION 7] LDAPResult
func (m ModifyResponse) size() int {
	return LDAPResult(m).sizeTagged(TagModifyResponse)
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
func (a AddRequest) size() (size int) {
	size += a.entry.size()
	size += a.attributes.size()
	size += sizeTagAndLength(TagAddRequest, size)
	return
}

//
//        AttributeList ::= SEQUENCE OF attribute Attribute
func (a AttributeList) size() (size int) {
	for _, att := range a {
		size += att.size()
	}
	size += sizeTagAndLength(tagSequence, size)
	return
}

//
//        AddResponse ::= [APPLICATION 9] LDAPResult
func (a AddResponse) size() int {
	return LDAPResult(a).sizeTagged(TagAddResponse)
}

//
//        DelRequest ::= [APPLICATION 10] LDAPDN
func (d DelRequest) size() int {
	return LDAPDN(d).sizeTagged(TagDelRequest)
}

//
//        DelResponse ::= [APPLICATION 11] LDAPResult
func (d DelResponse) size() int {
	return LDAPResult(d).sizeTagged(TagDelResponse)
}

//
//        ModifyDNRequest ::= [APPLICATION 12] SEQUENCE {
//             entry           LDAPDN,
//             newrdn          RelativeLDAPDN,
//             deleteoldrdn    BOOLEAN,
//             newSuperior     [0] LDAPDN OPTIONAL }
func (m ModifyDNRequest) size() (size int) {
	size += m.entry.size()
	size += m.newrdn.size()
	size += m.deleteoldrdn.size()
	if m.newSuperior != nil {
		size += m.newSuperior.sizeTagged(TagModifyDNRequestNewSuperior)
	}
	size += sizeTagAndLength(TagModifyDNRequest, size)
	return
}

//
//        ModifyDNResponse ::= [APPLICATION 13] LDAPResult
func (m ModifyDNResponse) size() int {
	return LDAPResult(m).sizeTagged(TagModifyDNResponse)
}

//
//        CompareRequest ::= [APPLICATION 14] SEQUENCE {
//             entry           LDAPDN,
//             ava             AttributeValueAssertion }
func (c CompareRequest) size() (size int) {
	size += c.entry.size()
	size += c.ava.size()
	size += sizeTagAndLength(TagCompareRequest, size)
	return
}

//
//        CompareResponse ::= [APPLICATION 15] LDAPResult
func (c CompareResponse) size() int {
	return LDAPResult(c).sizeTagged(TagCompareResponse)
}

//
//        AbandonRequest ::= [APPLICATION 16] MessageID
func (a AbandonRequest) size() int {
	return MessageID(a).sizeTagged(TagAbandonRequest)
}

//
//        ExtendedRequest ::= [APPLICATION 23] SEQUENCE {
//             requestName      [0] LDAPOID,
//             requestValue     [1] OCTET STRING OPTIONAL }
func (e ExtendedRequest) size() (size int) {
	size += e.requestName.sizeTagged(TagExtendedRequestName)
	if e.requestValue != nil {
		size += e.requestValue.sizeTagged(TagExtendedRequestValue)
	}
	size += sizeTagAndLength(TagExtendedRequest, size)
	return
}

//
//        ExtendedResponse ::= [APPLICATION 24] SEQUENCE {
//             COMPONENTS OF LDAPResult,
//             responseName     [10] LDAPOID OPTIONAL,
//             responseValue    [11] OCTET STRING OPTIONAL }
func (e ExtendedResponse) size() (size int) {
	size += e.LDAPResult.sizeComponents()
	if e.responseName != nil {
		size += e.responseName.sizeTagged(TagExtendedResponseName)
	}
	if e.responseValue != nil {
		size += e.responseValue.sizeTagged(TagExtendedResponseValue)
	}
	size += sizeTagAndLength(TagExtendedResponse, size)
	return
}

//
//        IntermediateResponse ::= [APPLICATION 25] SEQUENCE {
//             responseName     [0] LDAPOID OPTIONAL,
//             responseValue    [1] OCTET STRING OPTIONAL }
func (i IntermediateResponse) size() (size int) {
	if i.responseName != nil {
		size += i.responseName.sizeTagged(TagIntermediateResponseName)
	}
	if i.responseValue != nil {
		size += i.responseValue.sizeTagged(TagIntermediateResponseValue)
	}
	size += sizeTagAndLength(TagIntermediateResponse, size)
	return
}

//
//        END
//
