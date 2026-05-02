package types

// NamespaceName represents a namespace identifier
type NamespaceName string

// StringPtr returns a pointer to the string representation of NamespaceName, or nil if empty
func (n NamespaceName) StringPtr() *string {
	if n == "" {
		return nil
	}
	s := string(n)
	return &s
}

// NamespaceNamePtrToStringPtr converts *NamespaceName to *string
func NamespaceNamePtrToStringPtr(n *NamespaceName) *string {
	if n == nil {
		return nil
	}
	return (*n).StringPtr()
}

// ModuleName represents a module identifier
type ModuleName string

// StringPtr returns a pointer to the string representation of ModuleName, or nil if empty
func (m ModuleName) StringPtr() *string {
	if m == "" {
		return nil
	}
	s := string(m)
	return &s
}

// ModuleNamePtrToStringPtr converts *ModuleName to *string
func ModuleNamePtrToStringPtr(m *ModuleName) *string {
	if m == nil {
		return nil
	}
	return (*m).StringPtr()
}

// ModuleProviderName represents a provider identifier
type ModuleProviderName string

// StringPtr returns a pointer to the string representation of ModuleProviderName, or nil if empty
func (p ModuleProviderName) StringPtr() *string {
	if p == "" {
		return nil
	}
	s := string(p)
	return &s
}

// ModuleProviderNamePtrToStringPtr converts *ModuleProviderName to *string
func ModuleProviderNamePtrToStringPtr(p *ModuleProviderName) *string {
	if p == nil {
		return nil
	}
	return (*p).StringPtr()
}

// ModuleVersion represents version of module provider, e.g. "1.2.3"
type ModuleVersion string

// StringPtr returns a pointer to the string representation of ModuleVersion, or nil if empty
func (v ModuleVersion) StringPtr() *string {
	if v == "" {
		return nil
	}
	s := string(v)
	return &s
}

// ModuleVersionPtrToStringPtr converts *ModuleVersion to *string
func ModuleVersionPtrToStringPtr(v *ModuleVersion) *string {
	if v == nil {
		return nil
	}
	return (*v).StringPtr()
}

// ModuleProviderFrontendId represents Frontend ID for module provider, e.g. namespace/module/provider
type ModuleProviderFrontendId string

// ModuleProviderVersionFrontendId represents Frontend ID for module provider version, e.g. namespace/module/provider/1.2.3
type ModuleProviderVersionFrontendId string
