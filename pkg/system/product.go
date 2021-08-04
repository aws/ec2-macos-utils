package system

import (
	"fmt"
	"strings"
)

// Product defines the functionality necessary to create specific product types
type Product interface {
	// Name returns the product name (e.g. "Big Sur")
	Name() string

	// Version returns the product's release version (e.g. "11")
	Version() string

	// AltVersion returns the product's alternative version (e.g. "10.16").
	//
	// This is necessary since Big Sur will differentiate what version is returned depending on if the
	// SYSTEM_VERSION_COMPAT environment variable is set or not.
	AltVersion() string

	// String provides the implementation for the Stringer interface by formatting the Product's information
	String() string

	// Contains checks if a given product version string belongs to the Product's Version or AltVersion
	Contains(version string) bool
}

var (
	// Mojave represents the macOS release of version 10.14 aka "Mojave"
	Mojave Product = &ProductMojave{&baseProduct{"Mojave", "10.14", ""}}

	// Catalina represents the macOS release of version 10.15 aka "Catalina"
	Catalina Product = &ProductCatalina{&baseProduct{"Catalina", "10.15", ""}}

	// BigSur represents the macOS release of version 11 (10.16) aka "Big Sur"
	BigSur Product = &ProductBigSur{&baseProduct{"Big Sur", "11", "10.16"}}

	// products is a slice of Product(s) used to keep track of all known Product versions
	products []Product
)

func init() {
	products = append(products, Mojave, Catalina, BigSur)
}

// ProductMojave is the type used to identify the Mojave Product
type ProductMojave struct {
	*baseProduct
}

// ProductCatalina is the type used to identify the Catalina Product
type ProductCatalina struct {
	*baseProduct
}

// ProductBigSur is the type used to identify the BigSur Product
type ProductBigSur struct {
	*baseProduct
}

// baseProduct declares a standard type for referencing macOS releases
type baseProduct struct {
	name       string
	version    string
	altVersion string
}

// Name returns the baseProduct's associated name (e.g. "Big Sur")
func (p *baseProduct) Name() string {
	return p.name
}

// Version returns the baseProduct's version string (e.g. 11.4)
func (p *baseProduct) Version() string {
	return p.version
}

// AltVersion returns the baseProduct's altVersion string (e.g. 10.16)
func (p *baseProduct) AltVersion() string {
	return p.altVersion
}

// String implements the Stringer interface.
func (p *baseProduct) String() string {
	return fmt.Sprintf("%s %s", p.name, p.version)
}

// Contains checks to see if the baseProduct is the parent to a given version string (e.g. "11" is the parent of "11.4")
func (p *baseProduct) Contains(version string) bool {
	if p.altVersion != "" {
		return strings.HasPrefix(version, p.version) || strings.HasPrefix(version, p.altVersion)
	}

	return strings.HasPrefix(version, p.version)
}

// ProductFromVersion determines the correct baseProduct to return by comparing the given version with known baseProduct
// release versions and major versions.
func ProductFromVersion(version string) (Product, error) {
	for _, product := range products {
		if product.Contains(version) {
			return product, nil
		}
	}

	return nil, fmt.Errorf("product version [%s] does not match known versions", version)
}
