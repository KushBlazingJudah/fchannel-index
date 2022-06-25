package activitypub

import (
	"html/template"
	"time"
)

type AtContext struct {
	Context string `json:"@context,omitempty"`
}

type Collection struct {
	AtContext
	CollectionBase
}

type CollectionBase struct {
	Actor        *Actor       `json:"actor,omitempty"`
	Summary      string       `json:"summary,omitempty"`
	Type         string       `json:"type,omitempty"`
	TotalItems   int          `json:"totalItems,omitempty"`
	TotalImgs    int          `json:"totalImgs,omitempty"`
	OrderedItems []ObjectBase `json:"orderedItems,omitempty"`
	Items        []ObjectBase `json:"items,omitempty"`
}

type Actor struct {
	Type              string       `json:"type,omitempty"`
	Id                string       `json:"id,omitempty"`
	Inbox             string       `json:"inbox,omitempty"`
	Outbox            string       `json:"outbox,omitempty"`
	Following         string       `json:"following,omitempty"`
	Followers         string       `json:"followers,omitempty"`
	Name              string       `json:"name,omitempty"`
	PreferredUsername string       `json:"preferredUsername,omitempty"`
	PublicKey         PublicKeyPem `json:"publicKey,omitempty"`
	Summary           string       `json:"summary,omitempty"`
	AuthRequirement   []string     `json:"authrequirement,omitempty"`
	Restricted        bool         `json:"restricted"`
}

type PublicKeyPem struct {
	Id           string `json:"id,omitempty"`
	Owner        string `json:"owner,omitempty"`
	PublicKeyPem string `json:"publicKeyPem,omitempty"`
}

type ObjectBase struct {
	Type         string            `json:"type,omitempty"`
	Id           string            `json:"id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Option       []string          `json:"option,omitempty"`
	Alias        string            `json:"alias,omitempty"`
	AttributedTo string            `json:"attributedTo,omitempty"`
	TripCode     string            `json:"tripcode,omitempty"`
	Actor        string            `json:"actor,omitempty"`
	Audience     string            `json:"audience,omitempty"`
	ContentHTML  template.HTML     `json:"contenthtml,omitempty"`
	Content      string            `json:"content,omitempty"`
	EndTime      string            `json:"endTime,omitempty"`
	Generator    string            `json:"generator,omitempty"`
	Icon         string            `json:"icon,omitempty"`
	Image        string            `json:"image,omitempty"`
	InReplyTo    []ObjectBase      `json:"inReplyTo,omitempty"`
	Location     string            `json:"location,omitempty"`
	Preview      *NestedObjectBase `json:"preview,omitempty"`
	Published    time.Time         `json:"published,omitempty"`
	Updated      time.Time         `json:"updated,omitempty"`
	Object       *NestedObjectBase `json:"object,omitempty"`
	Attachment   []ObjectBase      `json:"attachment,omitempty"`
	Replies      *CollectionBase   `json:"replies,omitempty"`
	StartTime    string            `json:"startTime,omitempty"`
	Summary      string            `json:"summary,omitempty"`
	Tag          []ObjectBase      `json:"tag,omitempty"`
	Wallet       []CryptoCur       `json:"wallet,omitempty"`
	Deleted      string            `json:"deleted,omitempty"`
	Url          []ObjectBase      `json:"url,omitempty"`
	Href         string            `json:"href,omitempty"`
	To           []string          `json:"to,omitempty"`
	Bto          []string          `json:"bto,omitempty"`
	Cc           []string          `json:"cc,omitempty"`
	Bcc          string            `json:"Bcc,omitempty"`
	MediaType    string            `json:"mediatype,omitempty"`
	Duration     string            `json:"duration,omitempty"`
	Size         int64             `json:"size,omitempty"`
	Sensitive    bool              `json:"sensitive,omitempty"`
	Sticky       bool
	Locked       bool
}

type CryptoCur struct {
	Type    string `json:"type,omitempty"`
	Address string `json:"address,omitempty"`
}

type NestedObjectBase struct {
	AtContext
	Type         string          `json:"type,omitempty"`
	Id           string          `json:"id,omitempty"`
	Name         string          `json:"name,omitempty"`
	Alias        string          `json:"alias,omitempty"`
	AttributedTo string          `json:"attributedTo,omitempty"`
	TripCode     string          `json:"tripcode,omitempty"`
	Actor        string          `json:"actor,omitempty"`
	Audience     string          `json:"audience,omitempty"`
	ContentHTML  template.HTML   `json:"contenthtml,omitempty"`
	Content      string          `json:"content,omitempty"`
	EndTime      string          `json:"endTime,omitempty"`
	Generator    string          `json:"generator,omitempty"`
	Icon         string          `json:"icon,omitempty"`
	Image        string          `json:"image,omitempty"`
	InReplyTo    []ObjectBase    `json:"inReplyTo,omitempty"`
	Location     string          `json:"location,omitempty"`
	Preview      ObjectBase      `json:"preview,omitempty"`
	Published    time.Time       `json:"published,omitempty"`
	Attachment   []ObjectBase    `json:"attachment,omitempty"`
	Replies      *CollectionBase `json:"replies,omitempty"`
	StartTime    string          `json:"startTime,omitempty"`
	Summary      string          `json:"summary,omitempty"`
	Tag          []ObjectBase    `json:"tag,omitempty"`
	Updated      time.Time       `json:"updated,omitempty"`
	Deleted      string          `json:"deleted,omitempty"`
	Url          []ObjectBase    `json:"url,omitempty"`
	Href         string          `json:"href,omitempty"`
	To           []string        `json:"to,omitempty"`
	Bto          []string        `json:"bto,omitempty"`
	Cc           []string        `json:"cc,omitempty"`
	Bcc          string          `json:"Bcc,omitempty"`
	MediaType    string          `json:"mediatype,omitempty"`
	Duration     string          `json:"duration,omitempty"`
	Size         int64           `json:"size,omitempty"`
}
