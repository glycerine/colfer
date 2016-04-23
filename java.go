package colfer

import (
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateJava writes the code into the respective ".java" files.
func GenerateJava(basedir string, structs []*Struct) error {
	t := template.New("java-code").Delims("<:", ":>")
	template.Must(t.Parse(javaCode))
	template.Must(t.New("marshal").Parse(javaMarshal))
	template.Must(t.New("unmarshal").Parse(javaUnmarshal))

	for _, s := range structs {
		s.Pkg.NameNative = strings.Replace(s.Pkg.Name, "/", ".", -1)
	}

	for _, s := range structs {
		for _, f := range s.Fields {
			switch f.Type {
			default:
				if f.TypeRef == nil {
					f.TypeNative = f.Type
				} else {
					f.TypeNative = f.TypeRef.NameTitle()
					if pkg := f.TypeRef.Pkg.NameNative; pkg != s.Pkg.Name {
						f.TypeNative = pkg + "." + f.TypeNative
					}
				}
			case "bool":
				f.TypeNative = "boolean"
			case "uint32", "int32":
				f.TypeNative = "int"
			case "uint64", "int64":
				f.TypeNative = "long"
			case "float32":
				f.TypeNative = "float"
			case "float64":
				f.TypeNative = "double"
			case "timestamp":
				f.TypeNative = "java.time.Instant"
			case "text":
				f.TypeNative = "String"
			case "binary":
				f.TypeNative = "byte[]"
			}
		}

		pkgdir, err := MakePkgDir(&s.Pkg, basedir)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(pkgdir, s.NameTitle()+".java"))
		if err != nil {
			return err
		}
		defer f.Close()

		if err := t.Execute(f, s); err != nil {
			return err
		}
	}
	return nil
}

const javaCode = `package <:.Pkg.NameNative:>;

// This file was generated by colf(1); DO NOT EDIT


/**
 * @author Commander Colfer
 * @see <a href="https://github.com/pascaldekloe/colfer">Colfer's home</a>
 */
public class <:.NameTitle:> implements java.io.Serializable {

	private static final java.nio.charset.Charset utf8 = java.nio.charset.Charset.forName("UTF-8");

<:range .Fields:>	public <:.TypeNative:> <:.Name:>;
<:end:>

<:template "marshal" .:>
<:template "unmarshal" .:>
<:range .Fields:>	public <:.TypeNative:> get<:.NameTitle:>() {
		return this.<:.Name:>;
	}

	public void set<:.NameTitle:>(<:.TypeNative:> value) {
		this.<:.Name:> = value;
	}

<:end:>	/**
	 * Serializes an integer.
	 * @param buf the data destination.
	 * @param x the value.
	 */
	private static void putVarint(java.nio.ByteBuffer buf, int x) {
		while ((x & 0xffffff80) != 0) {
			buf.put((byte) (x | 0x80));
			x >>>= 7;
		}
		buf.put((byte) x);
	}

	/**
	 * Serializes an integer.
	 * @param buf the data destination.
	 * @param x the value.
	 */
	private static void putVarint(java.nio.ByteBuffer buf, long x) {
		while ((x & 0xffffffffffffff80L) != 0) {
			buf.put((byte) (x | 0x80));
			x >>>= 7;
		}
		buf.put((byte) x);
	}

	/**
	 * Deserializes a 32-bit integer.
	 * @param buf the data source.
	 * @return the value.
	 */
	private static int getVarint32(java.nio.ByteBuffer buf) {
		int x = 0;
		for (int shift = 0; shift != 28; shift += 7) {
			int b = buf.get() & 0xff;
			x |= (b & 0x7f) << shift;
			if (b < 0x80) return x;
		}
		long b = buf.get() & 0xffL;
		x |= b << 28;
		return x;
	}

	/**
	 * Deserializes a 64-bit integer.
	 * @param buf the data source.
	 * @return the value.
	 */
	private static long getVarint64(java.nio.ByteBuffer buf) {
		long x = 0;
		for (int shift = 0; shift != 63; shift += 7) {
			long b = buf.get() & 0xffL;
			x |= (b & 0x7f) << shift;
			if (b < 0x80) return x;
		}
		buf.get();
		x |= 1L << 63;
		return x;
	}

}
`

const javaMarshal = `	/**
	 * Writes in Colfer format.
	 * @param buf the data destination.
	 * @throws java.nio.BufferOverflowException when {@code buf} is too small.
	 */
	public final void marshal(java.nio.ByteBuffer buf) {
		buf.order(java.nio.ByteOrder.BIG_ENDIAN);
		buf.put((byte) 0x80);
<:range .Fields:><:if eq .Type "bool":>
		if (this.<:.Name:>) {
			buf.put((byte) <:.Index:>);
		}
<:else if eq .Type "uint32" "uint64":>
		if (this.<:.Name:> != 0) {
			buf.put((byte) <:.Index:>);
			putVarint(buf, this.<:.Name:>);
		}
<:else if eq .Type "int32":>
		if (this.<:.Name:> != 0) {
			int x = this.<:.Name:>;
			if (x < 0) {
				x = -x;
				buf.put((byte) (<:.Index:> | 0x80));
			} else
				buf.put((byte) <:.Index:>);
			putVarint(buf, x);
		}
<:else if eq .Type "int64":>
		if (this.<:.Name:> != 0) {
			long x = this.<:.Name:>;
			if (x < 0) {
				x = -x;
				buf.put((byte) (<:.Index:> | 0x80));
			} else
				buf.put((byte) <:.Index:>);
			putVarint(buf, x);
		}
<:else if eq .Type "float32":>
		if (this.<:.Name:> != 0.0f) {
			buf.put((byte) <:.Index:>);
			buf.putFloat(this.<:.Name:>);
		}
<:else if eq .Type "float64":>
		if (this.<:.Name:> != 0.0) {
			buf.put((byte) <:.Index:>);
			buf.putDouble(this.<:.Name:>);
		}
<:else if eq .Type "timestamp":>
		if (this.<:.Name:> != null) {
			long s = this.<:.Name:>.getEpochSecond();
			int ns = this.<:.Name:>.getNano();
			if (ns == 0) {
				if (s != 0) {
					buf.put((byte) <:.Index:>);
					buf.putLong(s);
				}
			} else {
				buf.put((byte) (<:.Index:> | 0x80));
				buf.putLong(s);
				buf.putInt(ns);
			}
		}
<:else if eq .Type "text":>
		if (this.<:.Name:> != null && ! this.<:.Name:>.isEmpty()) {
			java.nio.ByteBuffer bytes = utf8.encode(this.<:.Name:>);
			buf.put((byte) <:.Index:>);
			putVarint(buf, bytes.limit());
			buf.put(bytes);
		}
<:else if eq .Type "binary":>
		if (this.<:.Name:> != null && this.<:.Name:>.length != 0) {
			buf.put((byte) <:.Index:>);
			putVarint(buf, this.<:.Name:>.length);
			buf.put(this.<:.Name:>);
		}
<:else:>
		if (this.<:.Name:> != null) {
			buf.put((byte) <:.Index:>);
			this.<:.Name:>.marshal(buf);
		}
<:end:><:end:>
		buf.put((byte) 0x7f);
	}
`


const javaUnmarshal = `	/**
	 * Reads in Colfer format.
	 * @param buf the data source.
	 * @throws java.nio.BufferUnderflowException when {@code buf} is incomplete.
	 * @throws java.util.InputMismatchException on malformed data.
	 */
	public final void unmarshal(java.nio.ByteBuffer buf) {
		if (buf.get() != (byte) 0x80)
			throw new java.util.InputMismatchException("unknown header at byte 0");

		byte header = buf.get();
<:range .Fields:><:if eq .Type "bool":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = true;
			header = buf.get();
		}
<:else if eq .Type "uint32":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = getVarint32(buf);
			header = buf.get();
		}
<:else if eq .Type "uint64":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = getVarint64(buf);
			header = buf.get();
		}
<:else if eq .Type "int32":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = getVarint32(buf);
			header = buf.get();
		} else if (header == (byte) (<:.Index:> | 0x80)) {
			this.<:.Name:> = (~getVarint32(buf)) + 1;
			header = buf.get();
		}
<:else if eq .Type "int64":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = getVarint64(buf);
			header = buf.get();
		} else if (header == (byte) (<:.Index:> | 0x80)) {
			this.<:.Name:> = (~getVarint64(buf)) + 1;
			header = buf.get();
		}
<:else if eq .Type "float32":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = buf.getFloat();
			header = buf.get();
		}
<:else if eq .Type "float64":>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = buf.getDouble();
			header = buf.get();
		}
<:else if eq .Type "timestamp":>
		if (header == (byte) <:.Index:>) {
			long s = buf.getLong();
			this.<:.Name:> = java.time.Instant.ofEpochSecond(s);
			header = buf.get();
		} else if (header == (byte) (<:.Index:> | 0x80)) {
			long s = buf.getLong();
			int ns = buf.getInt();
			this.<:.Name:> = java.time.Instant.ofEpochSecond(s, ns);
			header = buf.get();
		}
<:else if eq .Type "text":>
		if (header == (byte) <:.Index:>) {
			int length = getVarint32(buf);
			java.nio.ByteBuffer blob = java.nio.ByteBuffer.allocate(length);
			buf.get(blob.array());
			this.<:.Name:> = utf8.decode(blob).toString();
			header = buf.get();
		}
<:else if eq .Type "binary":>
		if (header == (byte) <:.Index:>) {
			int length = getVarint32(buf);
			this.<:.Name:> = new byte[length];
			buf.get(this.<:.Name:>);
			header = buf.get();
		}
<:else:>
		if (header == (byte) <:.Index:>) {
			this.<:.Name:> = new <:.TypeNative:>();
			this.<:.Name:>.unmarshal(buf);
			header = buf.get();
		}
<:end:><:end:>
		if (header != 0x7f)
			throw new java.util.InputMismatchException(String.format("colfer: unknown header at byte %d", buf.position() - 1));
	}
`
