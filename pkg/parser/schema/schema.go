package schema

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

/*
HCL/Terraform 표현식 타입 정리:

## 기본 스칼라 타입 (LiteralValueExpr):
- 숫자: 42, 3.14, 1e10
- 불린: true, false  
- null: null

## 문자열 관련:
- 템플릿 표현식 (TemplateExpr): "hello", "prefix-${var.name}-suffix"
  └─ 단순 문자열: "hello" -> TemplateExpr.Parts[0]이 LiteralValueExpr
  └─ 보간 문자열: "hello ${var.name}" -> Parts에 LiteralValueExpr + TemplateWrapExpr

## 변수 참조 및 네임스페이스 접근:
- 스코프 순회 (ScopeTraversalExpr): 점 표기법으로 값에 접근하는 모든 표현식
  └─ 단순 이름: source, version (따옴표 없는 단일 이름)
  └─ 변수 참조: var.name, local.value 
  └─ 리소스 참조: aws_instance.web.id, data.aws_vpc.main.cidr_block
  └─ 모듈 참조: module.network.vpc_id

## 복합 구조:
- 튜플/배열 (TupleConsExpr): [1, 2, 3], ["a", "b"]
- 객체/맵 (ObjectConsExpr): {key = "value", foo = "bar"}
  └─ 키 표현식 (ObjectConsKeyExpr): 객체의 키를 감싸는 래퍼
      └─ Wrapped가 ScopeTraversalExpr: source = "value"
      └─ Wrapped가 TemplateExpr: "source" = "value"

## 함수 및 연산:
- 함수 호출 (FunctionCallExpr): toset([1,2,3]), length(var.list)
- 조건 표현식 (ConditionalExpr): condition ? true_val : false_val
- 이항 연산 (BinaryOpExpr): a + b, a == b
- 단항 연산 (UnaryOpExpr): !condition, -number

## 고급 표현식:
- 인덱스 접근 (IndexExpr): var.list[0], var.map["key"]
- 속성 접근 (GetAttrExpr): object.attribute
- 스플랫 표현식 (SplatExpr): var.list[*].name
- For 표현식 (ForExpr): [for item in list : item.name]
- 괄호 표현식 (ParenthesesExpr): (expression)

## 예시:
variable "example" {
  type = string              # <- ScopeTraversalExpr (단순 이름 'string')
  default = "hello"          # <- TemplateExpr (따옴표 문자열)
  count = 42                 # <- LiteralValueExpr (숫자)
  enabled = true             # <- LiteralValueExpr (불린)
  tags = {                   # <- ObjectConsExpr
    Name = "test"            # <- Name: ScopeTraversalExpr (단순 이름), "test": TemplateExpr
    "Environment" = "dev"    # <- "Environment": TemplateExpr, "dev": TemplateExpr
  }
  computed = var.prefix      # <- ScopeTraversalExpr (변수 참조)
}
*/

type Block interface {
	Parse(file *hcl.File, block *hclsyntax.Block) error
}

func parseAttributeToInterface(file *hcl.File, attr *hclsyntax.Attribute) interface{} {
	//
	// Literal인 string, number, bool, null을 최대한 타입대로 반환
	// 예를 들어, variable block의 type에서의 string은 LiteralValue이고
	// deufalt에서의 "default"는 Template
	//
	if lv, ok := attr.Expr.(*hclsyntax.LiteralValueExpr); ok {
		switch lv.Val.Type() {
		case cty.String:
			// HCL의 string은 go의 string으로 변환
			return lv.Val.AsString()
		case cty.Number:
			if lv.Val.AsBigFloat().IsInt() {
				// HCL의 nubmer가 정수면 go의 int64로 변환
				i, _ := lv.Val.AsBigFloat().Int64()
				return i
			}

			// HCL의 nubmer가 정수가 아니면 go의 float64로 변환
			f, _ := lv.Val.AsBigFloat().Float64()
			return f
		case cty.Bool:
			// HCL의 bool을 go의 bool로 변환
			return lv.Val.True()
		}
		if lv.Val.IsNull() {
			// HCL의 null을 go의 nil로 변환
			return nil
		}
	}

	// 템플릿 표현식에서 문자열 추출
	if te, ok := attr.Expr.(*hclsyntax.TemplateExpr); ok {
		if len(te.Parts) == 1 {
			if lv, ok := te.Parts[0].(*hclsyntax.LiteralValueExpr); ok && lv.Val.Type() == cty.String {
				return lv.Val.AsString()
			}
		}
	}

	// 복잡한 표현식의 경우 원본 HCL 구문 반환
	raw := attr.Expr.Range().SliceBytes(file.Bytes)
	return strings.TrimSpace(string(raw))
}

func parseAttributeToString(file *hcl.File, attr *hclsyntax.Attribute) string {
	value := parseAttributeToInterface(file, attr)
	if str, ok := value.(string); ok {
		return str
	}

	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseAttributeToBool(file *hcl.File, attr *hclsyntax.Attribute) bool {
	value := parseAttributeToInterface(file, attr)
	if boolVal, ok := value.(bool); ok {
		return boolVal
	}

	switch v := value.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "true":
			return true
		case "false":
			return false
		}
	default:
		return false
	}

	return false
}

// 배열 속성을 문자열 슬라이스로 파싱
func parseAttributeToStringList(file *hcl.File, attr *hclsyntax.Attribute) []string {
	// 튜플 표현식 처리 (HCL 배열)
	if tupleExpr, ok := attr.Expr.(*hclsyntax.TupleConsExpr); ok {
		result := make([]string, 0, len(tupleExpr.Exprs))
		for _, expr := range tupleExpr.Exprs {
			fakeAttr := &hclsyntax.Attribute{Expr: expr}
			result = append(result, parseAttributeToString(file, fakeAttr))
		}
		return result
	}

	// 복잡한 표현식의 경우 원본을 간단히 파싱
	raw := string(attr.Expr.Range().SliceBytes(file.Bytes))
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "[]")

	if raw == "" {
		return []string{}
	}

	// 간단한 쉼표 분할 (더 정교한 파싱 필요시 개선)
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, "\"")
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

// 객체 속성을 맵으로 파싱 (key-value 쌍)
func parseAttributeToStringMap(file *hcl.File, attr *hclsyntax.Attribute) map[string]string {
	result := make(map[string]string)

	// 객체 표현식 처리
	if objExpr, ok := attr.Expr.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range objExpr.Items {
			key := extractObjectKey(item.KeyExpr)
			if key != "" {
				fakeAttr := &hclsyntax.Attribute{Expr: item.ValueExpr}
				value := parseAttributeToString(file, fakeAttr)
				result[key] = value
			}
		}
	}

	return result
}

// 객체 키를 추출하는 헬퍼 함수  
func extractObjectKey(keyExpr hclsyntax.Expression) string {
	switch key := keyExpr.(type) {
	case *hclsyntax.ObjectConsKeyExpr:
		if key.Wrapped != nil {
			if literalKey, ok := key.Wrapped.(*hclsyntax.LiteralValueExpr); ok {
				return literalKey.Val.AsString()
			}
			if scopeKey, ok := key.Wrapped.(*hclsyntax.ScopeTraversalExpr); ok {
				return scopeKey.Traversal.RootName()
			}
		}
		return ""
	case *hclsyntax.LiteralValueExpr:
		return key.Val.AsString()
	case *hclsyntax.ScopeTraversalExpr:
		// 식별자인 경우 (예: source, version)
		if len(key.Traversal) > 0 {
			if rootName := key.Traversal.RootName(); rootName != "" {
				return rootName
			}
		}
	}
	return ""
}
