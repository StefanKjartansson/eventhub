package eventhub

type Query struct {
	Origin   string   `url:"origin,omitempty" schema:"origin"`
	Key      string   `url:"key,omitempty" schema:"key"`
	Entities []string `url:"entities,omitempty" schema:"entities"`
	Actors   []string `url:"actors,omitempty" schema:"actors"`
}
