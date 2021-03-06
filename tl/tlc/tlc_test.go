package tlc

import (
	"github.com/PROger4ever/diff"
	"github.com/PROger4ever/telegramapi/tl/knownschemas"
	"github.com/PROger4ever/telegramapi/tl/tlschema"
	"testing"
)

func TestSimple(t *testing.T) {
	sch := tlschema.MustParse(`
        nearestDc#8e1a1775 country:string this_dc:int nearest_dc:int = NearestDc;
        --- functions ---
        help.getNearestDc#1fb33026 = NearestDc;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLNearestDC struct {
            Country   string
            ThisDC    int   
            NearestDC int   
        }

        func (o *TLNearestDC) Cmd() uint32 {
            return TagNearestDC
        }

        func (o *TLNearestDC) ReadBareFrom(r *tl.Reader) {
            o.Country = r.ReadString()
            o.ThisDC = r.ReadInt()
            o.NearestDC = r.ReadInt()
        }

        func (o *TLNearestDC) WriteBareTo(w *tl.Writer) {
            w.WriteString(o.Country)
            w.WriteInt(o.ThisDC)
            w.WriteInt(o.NearestDC)
        }

        type TLHelpGetNearestDC struct {
        }

        func (o *TLHelpGetNearestDC) Cmd() uint32 {
            return TagHelpGetNearestDC
        }

        func (o *TLHelpGetNearestDC) ReadBareFrom(r *tl.Reader) {
        }

        func (o *TLHelpGetNearestDC) WriteBareTo(w *tl.Writer) {
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:int = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFoo struct {
            Bar int
        }

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            o.Bar = r.ReadInt()
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(o.Bar)
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBigInt(t *testing.T) {
	sch := tlschema.MustParse(`
        resPQ#11223344 pq:bytes = ResPQ;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLResPQ struct {
            PQ *big.Int
        }

        func (o *TLResPQ) Cmd() uint32 {
            return TagResPQ
        }

        func (o *TLResPQ) ReadBareFrom(r *tl.Reader) {
            o.PQ = r.ReadBigInt()
        }

        func (o *TLResPQ) WriteBareTo(w *tl.Writer) {
            w.WriteBigInt(o.PQ)
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestVectorBareInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:Vector<int> = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFoo struct {
            Bar []int
        }

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            if cmd := r.ReadCmd(); cmd != TagVector {
                r.Fail(errors.New("expected: vector"))
            }
            o.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(o.Bar); i++ {
                o.Bar[i] = r.ReadInt()
            }
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteCmd(TagVector)
            w.WriteInt(len(o.Bar))
            for i := 0; i < len(o.Bar); i++ {
                w.WriteInt(o.Bar[i])
            }
        }        
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBareInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:%Vector<int> = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFoo struct {
            Bar []int
        }

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            o.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(o.Bar); i++ {
                o.Bar[i] = r.ReadInt()
            }
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(o.Bar))
            for i := 0; i < len(o.Bar); i++ {
                w.WriteInt(o.Bar[i])
            }
        }        
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBareStruct(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:vector<%Boz> = Foo;
        boz#99887766 = Boz;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFoo struct {
            Bar []*TLBoz
        }

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            o.Bar = make([]*TLBoz, r.ReadInt())
            for i := 0; i < len(o.Bar); i++ {
                o.Bar[i] = new(TLBoz)
                o.Bar[i].ReadBareFrom(r)
            }
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(o.Bar))
            for i := 0; i < len(o.Bar); i++ {
                o.Bar[i].WriteBareTo(w)
            }
        }   

        type TLBoz struct {
        }

        func (o *TLBoz) Cmd() uint32 {
            return TagBoz
        }

        func (o *TLBoz) ReadBareFrom(r *tl.Reader) {
        }

        func (o *TLBoz) WriteBareTo(w *tl.Writer) {
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBoxedStruct(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:vector<Boz> = Foo;
        boz#99887766 = Boz;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFoo struct {
            Bar []*TLBoz
        }

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            o.Bar = make([]*TLBoz, r.ReadInt())
            for i := 0; i < len(o.Bar); i++ {
                if cmd := r.ReadCmd(); cmd != TagBoz {
                    r.Fail(errors.New("expected: boz"))
                }
                o.Bar[i] = new(TLBoz)
                o.Bar[i].ReadBareFrom(r)
            }
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(o.Bar))
            for i := 0; i < len(o.Bar); i++ {
                w.WriteCmd(TagBoz)
                o.Bar[i].WriteBareTo(w)
            }
        }   

        type TLBoz struct {
        }

        func (o *TLBoz) Cmd() uint32 {
            return TagBoz
        }

        func (o *TLBoz) ReadBareFrom(r *tl.Reader) {
        }

        func (o *TLBoz) WriteBareTo(w *tl.Writer) {
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestMultiCtorType(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 x:int = Foo;
        bar#99887766 y:string = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLFooType interface {
            IsTLFoo()
            Cmd() uint32
            ReadBareFrom(r *tl.Reader)
            WriteBareTo(w *tl.Writer)
        }

        type TLFoo struct {
            X int
        }

        func (o *TLFoo) IsTLFoo() {}

        func (o *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (o *TLFoo) ReadBareFrom(r *tl.Reader) {
            o.X = r.ReadInt()
        }

        func (o *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(o.X)
        }

        type TLBar struct {
            Y string
        }

        func (o *TLBar) IsTLFoo() {}

        func (o *TLBar) Cmd() uint32 {
            return TagBar
        }

        func (o *TLBar) ReadBareFrom(r *tl.Reader) {
            o.Y = r.ReadString()
        }

        func (o *TLBar) WriteBareTo(w *tl.Writer) {
            w.WriteString(o.Y)
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestConditionalFlagField(t *testing.T) {
	sch := tlschema.MustParse(`
        dcOption#5d8c6cc flags:# ipv6:flags.0?true media_only:flags.1?true tcpo_only:flags.2?true id:int ip_address:string port:flags.3?int = DcOption;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLDCOption struct {
            Flags     uint
            ID        int
            IPAddress string
            Port      int
        }

        func (o *TLDCOption) Cmd() uint32 {
            return TagDCOption
        }

        func (o *TLDCOption) ReadBareFrom(r *tl.Reader) {
            o.Flags = uint(r.ReadUint32())
            o.ID = r.ReadInt()
            o.IPAddress = r.ReadString()
            if (o.Flags & (1 << 3)) != 0 {
                o.Port = r.ReadInt()
            }
        }

        func (o *TLDCOption) WriteBareTo(w *tl.Writer) {
            w.WriteUint32(uint32(o.Flags))
            w.WriteInt(o.ID)
            w.WriteString(o.IPAddress)
            if (o.Flags & (1 << 3)) != 0 {
                w.WriteInt(o.Port)
            }
        }

        func (o *TLDCOption) IPv6() bool {
            return (o.Flags & (1 << 0)) != 0
        }

        func (o *TLDCOption) SetIPv6(v bool) {
            if v {
                o.Flags |= (1 << 0)
            } else {
                o.Flags &= ^uint(1 << 0)
            }
        }

        func (o *TLDCOption) MediaOnly() bool {
            return (o.Flags & (1 << 1)) != 0
        }

        func (o *TLDCOption) SetMediaOnly(v bool) {
            if v {
                o.Flags |= (1 << 1)
            } else {
                o.Flags &= ^uint(1 << 1)
            }
        }

        func (o *TLDCOption) TCPoOnly() bool {
            return (o.Flags & (1 << 2)) != 0
        }

        func (o *TLDCOption) SetTCPoOnly(v bool) {
            if v {
                o.Flags |= (1 << 2)
            } else {
                o.Flags &= ^uint(1 << 2)
            }
        }

        func (o *TLDCOption) HasPort() bool {
            return (o.Flags & (1 << 3)) != 0
        }

        func (o *TLDCOption) SetHasPort(v bool) {
            if v {
                o.Flags |= (1 << 3)
            } else {
                o.Flags &= ^uint(1 << 3)
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestConditionalConflict(t *testing.T) {
	sch := tlschema.MustParse(`
        messages.botCallbackAnswer#36585ea4 flags:# has_url:flags.3?true url:flags.2?string = messages.BotCallbackAnswer;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        type TLMessagesBotCallbackAnswer struct {
            Flags uint
            URL   string
        }

        func (o *TLMessagesBotCallbackAnswer) Cmd() uint32 {
            return TagMessagesBotCallbackAnswer
        }

        func (o *TLMessagesBotCallbackAnswer) ReadBareFrom(r *tl.Reader) {
            o.Flags = uint(r.ReadUint32())
            if (o.Flags & (1 << 2)) != 0 {
                o.URL = r.ReadString()
            }
        }

        func (o *TLMessagesBotCallbackAnswer) WriteBareTo(w *tl.Writer) {
            w.WriteUint32(uint32(o.Flags))
            if (o.Flags & (1 << 2)) != 0 {
                w.WriteString(o.URL)
            }
        }

        func (o *TLMessagesBotCallbackAnswer) HasURL() bool {
            return (o.Flags & (1 << 3)) != 0
        }

        func (o *TLMessagesBotCallbackAnswer) SetHasURL(v bool) {
            if v {
                o.Flags |= (1 << 3)
            } else {
                o.Flags &= ^uint(1 << 3)
            }
        }

        func (o *TLMessagesBotCallbackAnswer) HasURLField() bool {
            return (o.Flags & (1 << 2)) != 0
        }

        func (o *TLMessagesBotCallbackAnswer) SetHasURLField(v bool) {
            if v {
                o.Flags |= (1 << 2)
            } else {
                o.Flags &= ^uint(1 << 2)
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestMTProto(t *testing.T) {
	sch := tlschema.MustParse(knownschemas.MTProtoSchema)
	GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
}

func TestTelegram(t *testing.T) {
	sch := tlschema.MustParse(knownschemas.TelegramSchema)
	GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
}
