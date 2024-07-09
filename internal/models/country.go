package models

type Country struct {
	Name   string `field:"name" json:"name"`
	Alpha2 string `field:"alpha2" json:"alpha2"`
	Alpha3 string `field:"alpha3" json:"alpha3"`
	Region string `field:"region" json:"region"`
}
