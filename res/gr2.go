package res

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"math"
	"strings"
)

const (
	gr2HeaderSize   = 0x20
	gr2SectorSize   = 44
	gr2FixupSize    = 12
	gr2OodleTailPad = 4
)

var (
	gr2MagicFF6LE = [4]uint32{0xCAB0_67B8, 0x0FB1_6DF8, 0x7E8C_7284, 0x1E00_195E}
	gr2MagicFF7LE = [4]uint32{0xC06C_DE29, 0x2B53_A4BA, 0xA5B7_F525, 0xEEE2_66F6}
)

type GR2SectorRef struct {
	Sector   uint32
	Position uint32
}

type GR2SectorInfo struct {
	CompressType  uint32
	DataOffset    uint32
	CompressedLen uint32
	DecompressLen uint32
	OodleStop0    uint32
	OodleStop1    uint32
	FixupOffset   uint32
	FixupCount    uint32
}

type GR2Container struct {
	Version       uint32
	Data          []byte
	SectorOffsets []int
	Sectors       []GR2SectorInfo
	TypeRef       GR2SectorRef
	RootRef       GR2SectorRef
}

func ParseGR2Container(data []byte) (*GR2Container, error) {
	if len(data) < 0x44 {
		return nil, fmt.Errorf("gr2: unexpected EOF")
	}
	magic := [4]uint32{
		readGR2U32At(data, 0),
		readGR2U32At(data, 4),
		readGR2U32At(data, 8),
		readGR2U32At(data, 12),
	}
	if magic != gr2MagicFF6LE && magic != gr2MagicFF7LE {
		return nil, fmt.Errorf("gr2: invalid magic")
	}
	if readGR2U32At(data, 0x18) != 0 {
		return nil, fmt.Errorf("gr2: invalid pointer format")
	}

	version := readGR2U32At(data, 0x20)
	if version != 6 && version != 7 {
		return nil, fmt.Errorf("gr2: unsupported version %d", version)
	}
	totalSize := int(readGR2U32At(data, 0x24))
	if totalSize != len(data) {
		return nil, fmt.Errorf("gr2: size mismatch")
	}
	fileCRC := readGR2U32At(data, 0x28)
	fileInfoSize := int(readGR2U32At(data, 0x2c))
	sectorCount := int(readGR2U32At(data, 0x30))
	typeRef := GR2SectorRef{
		Sector:   readGR2U32At(data, 0x34),
		Position: readGR2U32At(data, 0x38),
	}
	rootRef := GR2SectorRef{
		Sector:   readGR2U32At(data, 0x3c),
		Position: readGR2U32At(data, 0x40),
	}

	crcStart := gr2HeaderSize + fileInfoSize
	if crcStart < 0 || crcStart > len(data) {
		return nil, fmt.Errorf("gr2: bad crc offset")
	}
	if crc32.ChecksumIEEE(data[crcStart:]) != fileCRC {
		return nil, fmt.Errorf("gr2: crc mismatch")
	}

	sectorTable := gr2HeaderSize + fileInfoSize
	if sectorTable+sectorCount*gr2SectorSize > len(data) {
		return nil, fmt.Errorf("gr2: sector table truncated")
	}
	sectors := make([]GR2SectorInfo, 0, sectorCount)
	total := 0
	for i := 0; i < sectorCount; i++ {
		base := sectorTable + i*gr2SectorSize
		sector := GR2SectorInfo{
			CompressType:  readGR2U32At(data, base),
			DataOffset:    readGR2U32At(data, base+4),
			CompressedLen: readGR2U32At(data, base+8),
			DecompressLen: readGR2U32At(data, base+12),
			OodleStop0:    readGR2U32At(data, base+20),
			OodleStop1:    readGR2U32At(data, base+24),
			FixupOffset:   readGR2U32At(data, base+28),
			FixupCount:    readGR2U32At(data, base+32),
		}
		sectors = append(sectors, sector)
		total += int(sector.DecompressLen)
	}

	out := make([]byte, total)
	sectorOffsets := make([]int, 0, sectorCount)
	offset := 0
	for _, sector := range sectors {
		sectorOffsets = append(sectorOffsets, offset)
		dst := out[offset : offset+int(sector.DecompressLen)]
		srcStart := int(sector.DataOffset)
		if sector.CompressType == 0 {
			srcEnd := srcStart + int(sector.DecompressLen)
			if srcStart < 0 || srcEnd > len(data) {
				return nil, fmt.Errorf("gr2: uncompressed sector truncated")
			}
			copy(dst, data[srcStart:srcEnd])
		} else {
			srcEnd := srcStart + int(sector.CompressedLen)
			if srcStart < 0 || srcEnd > len(data) {
				return nil, fmt.Errorf("gr2: compressed sector truncated")
			}
			compressed := append([]byte(nil), data[srcStart:srcEnd]...)
			compressed = append(compressed, make([]byte, gr2OodleTailPad)...)
			if err := decompressGR2Oodle(compressed, dst, sector.OodleStop0, sector.OodleStop1); err != nil {
				return nil, err
			}
		}
		offset += int(sector.DecompressLen)
	}
	if err := applyGR2Fixups(data, out, sectorOffsets, sectors); err != nil {
		return nil, err
	}

	return &GR2Container{
		Version:       version,
		Data:          out,
		SectorOffsets: sectorOffsets,
		Sectors:       sectors,
		TypeRef:       typeRef,
		RootRef:       rootRef,
	}, nil
}

func (c *GR2Container) RefOffset(ref GR2SectorRef) (int, error) {
	if int(ref.Sector) < 0 || int(ref.Sector) >= len(c.SectorOffsets) {
		return 0, fmt.Errorf("gr2: bad sector ref %d", ref.Sector)
	}
	return c.SectorOffsets[ref.Sector] + int(ref.Position), nil
}

func applyGR2Fixups(file []byte, data []byte, sectorOffsets []int, sectors []GR2SectorInfo) error {
	for i, sector := range sectors {
		for k := 0; k < int(sector.FixupCount); k++ {
			base := int(sector.FixupOffset) + k*gr2FixupSize
			if base+gr2FixupSize > len(file) {
				return fmt.Errorf("gr2: fixup table truncated")
			}
			srcOffset := int(readGR2U32At(file, base))
			dstSector := int(readGR2U32At(file, base+4))
			dstOffset := int(readGR2U32At(file, base+8))
			if dstSector < 0 || dstSector >= len(sectorOffsets) {
				return fmt.Errorf("gr2: bad fixup sector")
			}
			target := uint32(sectorOffsets[dstSector] + dstOffset)
			at := sectorOffsets[i] + srcOffset
			if at < 0 || at+4 > len(data) {
				return fmt.Errorf("gr2: fixup target truncated")
			}
			binary.LittleEndian.PutUint32(data[at:at+4], target)
		}
	}
	return nil
}

func readGR2U32At(data []byte, off int) uint32 {
	return binary.LittleEndian.Uint32(data[off : off+4])
}

func readGR2U32(data []byte, off int) (uint32, error) {
	if off < 0 || off+4 > len(data) {
		return 0, fmt.Errorf("gr2: unexpected EOF")
	}
	return binary.LittleEndian.Uint32(data[off : off+4]), nil
}

func readGR2I32(data []byte, off int) (int32, error) {
	v, err := readGR2U32(data, off)
	return int32(v), err
}

func readGR2F32(data []byte, off int) (float32, error) {
	v, err := readGR2U32(data, off)
	return math.Float32frombits(v), err
}

type gr2MemberType uint32

const (
	gr2MemberEnd gr2MemberType = iota
	gr2MemberInline
	gr2MemberReference
	gr2MemberReferenceToArray
	gr2MemberArrayOfReferences
	gr2MemberVariantReference
	gr2MemberUnsupported
	gr2MemberReferenceToVariantArray
	gr2MemberString
	gr2MemberTransform
	gr2MemberReal32
	gr2MemberInt8
	gr2MemberUInt8
	gr2MemberBinormalInt8
	gr2MemberNormalUInt8
	gr2MemberInt16
	gr2MemberUInt16
	gr2MemberBinormalInt16
	gr2MemberNormalUInt16
	gr2MemberInt32
	gr2MemberUInt32
	gr2MemberReal16
	gr2MemberEmptyReference
)

const (
	gr2TypeDefSize  = 32
	gr2TransformLen = 68
)

type gr2RawMember struct {
	memberType gr2MemberType
	name       string
	refOff     int
	arrayWidth uint32
}

type gr2Field struct {
	name   string
	refOff int
	slot   int
}

type gr2Nav struct {
	data      []byte
	sizeCache map[int]int
}

func newGR2Nav(data []byte) gr2Nav {
	return gr2Nav{data: data, sizeCache: make(map[int]int)}
}

func (n *gr2Nav) u32(off int) (uint32, error) {
	return readGR2U32(n.data, off)
}

func (n *gr2Nav) i32(off int) (int32, error) {
	return readGR2I32(n.data, off)
}

func (n *gr2Nav) f32(off int) (float32, error) {
	return readGR2F32(n.data, off)
}

func (n *gr2Nav) cstr(off int) string {
	if off <= 0 || off >= len(n.data) {
		return ""
	}
	end := off
	for end < len(n.data) && n.data[end] != 0 {
		end++
	}
	if end >= len(n.data) {
		return ""
	}
	return string(n.data[off:end])
}

func (n *gr2Nav) members(typeOff int) ([]gr2RawMember, error) {
	var out []gr2RawMember
	for off := typeOff; ; off += gr2TypeDefSize {
		if off < 0 || off+gr2TypeDefSize > len(n.data) {
			return nil, fmt.Errorf("gr2: member table truncated")
		}
		rawType, err := n.u32(off)
		if err != nil {
			return nil, err
		}
		memberType := gr2MemberType(rawType)
		if memberType == gr2MemberEnd {
			break
		}
		if memberType > gr2MemberEmptyReference {
			return nil, fmt.Errorf("gr2: bad member type %d", rawType)
		}
		namePtr, err := n.u32(off + 4)
		if err != nil {
			return nil, err
		}
		refPtr, err := n.u32(off + 8)
		if err != nil {
			return nil, err
		}
		width, err := n.u32(off + 12)
		if err != nil {
			return nil, err
		}
		out = append(out, gr2RawMember{
			memberType: memberType,
			name:       n.cstr(int(namePtr)),
			refOff:     int(refPtr),
			arrayWidth: width,
		})
	}
	return out, nil
}

func (n *gr2Nav) typeSize(typeOff int) (int, error) {
	if size, ok := n.sizeCache[typeOff]; ok {
		return size, nil
	}
	n.sizeCache[typeOff] = 0
	members, err := n.members(typeOff)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, member := range members {
		size, err := n.memberSize(member)
		if err != nil {
			return 0, err
		}
		total += size
	}
	n.sizeCache[typeOff] = total
	return total, nil
}

func (n *gr2Nav) memberSize(member gr2RawMember) (int, error) {
	count := int(member.arrayWidth)
	if count < 1 {
		count = 1
	}
	switch member.memberType {
	case gr2MemberInline:
		size, err := n.typeSize(member.refOff)
		return size * count, err
	case gr2MemberReference, gr2MemberString, gr2MemberEmptyReference:
		return 4, nil
	case gr2MemberReferenceToArray, gr2MemberArrayOfReferences, gr2MemberVariantReference:
		return 8, nil
	case gr2MemberReferenceToVariantArray:
		return 12, nil
	case gr2MemberTransform:
		return gr2TransformLen, nil
	case gr2MemberReal32, gr2MemberInt32, gr2MemberUInt32:
		return 4 * count, nil
	case gr2MemberInt16, gr2MemberUInt16, gr2MemberBinormalInt16, gr2MemberNormalUInt16, gr2MemberReal16:
		return 2 * count, nil
	case gr2MemberInt8, gr2MemberUInt8, gr2MemberBinormalInt8, gr2MemberNormalUInt8:
		return count, nil
	default:
		return 0, nil
	}
}

func (n *gr2Nav) fields(typeOff, objOff int) ([]gr2Field, error) {
	members, err := n.members(typeOff)
	if err != nil {
		return nil, err
	}
	fields := make([]gr2Field, 0, len(members))
	slot := objOff
	for _, member := range members {
		size, err := n.memberSize(member)
		if err != nil {
			return nil, err
		}
		fields = append(fields, gr2Field{name: member.name, refOff: member.refOff, slot: slot})
		slot += size
	}
	return fields, nil
}

func (n *gr2Nav) array(slot int) (count int, ptr int, err error) {
	rawCount, err := n.i32(slot)
	if err != nil {
		return 0, 0, err
	}
	if rawCount < 0 {
		rawCount = 0
	}
	rawPtr, err := n.u32(slot + 4)
	if err != nil {
		return 0, 0, err
	}
	return int(rawCount), int(rawPtr), nil
}

func (n *gr2Nav) variantArray(slot int) (typ int, count int, ptr int, err error) {
	rawType, err := n.u32(slot)
	if err != nil {
		return 0, 0, 0, err
	}
	rawCount, err := n.i32(slot + 4)
	if err != nil {
		return 0, 0, 0, err
	}
	if rawCount < 0 {
		rawCount = 0
	}
	rawPtr, err := n.u32(slot + 8)
	if err != nil {
		return 0, 0, 0, err
	}
	return int(rawType), int(rawCount), int(rawPtr), nil
}

func gr2FindField(fields []gr2Field, name string) *gr2Field {
	for i := range fields {
		if fields[i].name == name {
			return &fields[i]
		}
	}
	return nil
}

func gr2I32Field(nav *gr2Nav, fields []gr2Field, name string) int32 {
	field := gr2FindField(fields, name)
	if field == nil {
		return 0
	}
	value, err := nav.i32(field.slot)
	if err != nil {
		return 0
	}
	return value
}

func gr2F32Field(nav *gr2Nav, fields []gr2Field, name string) float32 {
	field := gr2FindField(fields, name)
	if field == nil {
		return 0
	}
	value, err := nav.f32(field.slot)
	if err != nil {
		return 0
	}
	return value
}

func gr2StringField(nav *gr2Nav, fields []gr2Field, name string) string {
	field := gr2FindField(fields, name)
	if field == nil {
		return ""
	}
	ptr, err := nav.u32(field.slot)
	if err != nil {
		return ""
	}
	return nav.cstr(int(ptr))
}

type GR2Transform struct {
	Flags      uint32
	Position   [3]float32
	Rotation   [4]float32
	ScaleShear [9]float32
}

var GR2IdentityTransform = GR2Transform{
	Rotation:   [4]float32{0, 0, 0, 1},
	ScaleShear: [9]float32{1, 0, 0, 0, 1, 0, 0, 0, 1},
}

func readGR2Transform(nav *gr2Nav, slot int) (GR2Transform, error) {
	out := GR2Transform{}
	flags, err := nav.u32(slot)
	if err != nil {
		return out, err
	}
	out.Flags = flags
	for i := range out.Position {
		out.Position[i], err = nav.f32(slot + 4 + i*4)
		if err != nil {
			return out, err
		}
	}
	for i := range out.Rotation {
		out.Rotation[i], err = nav.f32(slot + 16 + i*4)
		if err != nil {
			return out, err
		}
	}
	for i := range out.ScaleShear {
		out.ScaleShear[i], err = nav.f32(slot + 32 + i*4)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

const (
	GR2TextureEncodingRaw  int32 = 1
	GR2TextureEncodingS3TC int32 = 2
	GR2TextureEncodingBink int32 = 3
)

type GR2Texture struct {
	FromFileName  string
	Width         int32
	Height        int32
	Encoding      int32
	SubFormat     int32
	BytesPerPixel int32
	ComponentBits [4]int32
	Pixels        []byte
}

func (t GR2Texture) HasAlpha() bool {
	return t.ComponentBits[3] != 0
}

func (t GR2Texture) RGBA() ([]byte, error) {
	width, height := int(t.Width), int(t.Height)
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("gr2 texture: invalid size %dx%d", width, height)
	}
	switch {
	case t.Encoding == GR2TextureEncodingRaw && t.BytesPerPixel == 4:
		expected := width * height * 4
		if len(t.Pixels) < expected {
			return nil, fmt.Errorf("gr2 texture: raw pixels truncated")
		}
		return append([]byte(nil), t.Pixels[:expected]...), nil
	case t.Encoding == GR2TextureEncodingBink:
		return decodeGR2Bink(t.Pixels, width, height, t.HasAlpha())
	default:
		return nil, fmt.Errorf("gr2 texture: unsupported encoding %d", t.Encoding)
	}
}

func (t GR2Texture) Image() (*image.RGBA, error) {
	rgba, err := t.RGBA()
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, int(t.Width), int(t.Height)))
	copy(img.Pix, rgba)
	return img, nil
}

type GR2Material struct {
	Name         string
	TextureIndex *int
}

type GR2Bone struct {
	Name         string
	ParentIndex  int32
	Transform    GR2Transform
	InverseWorld [16]float32
}

type GR2Skeleton struct {
	Name  string
	Bones []GR2Bone
}

type GR2Vertex struct {
	Position    [3]float32
	Normal      [3]float32
	UV          [2]float32
	BoneWeights [4]byte
	BoneIndices [4]byte
}

type GR2VertexData struct {
	Vertices []GR2Vertex
}

type GR2TriGroup struct {
	MaterialIndex int32
	TriFirst      int32
	TriCount      int32
}

type GR2TriTopology struct {
	Groups  []GR2TriGroup
	Indices []uint32
}

type GR2Mesh struct {
	Name            string
	VertexDataIndex *int
	TopologyIndex   *int
	MaterialIndices []int
	BoneBindings    []string
}

type GR2Model struct {
	Name             string
	SkeletonIndex    *int
	InitialPlacement GR2Transform
	MeshIndices      []int
}

type GR2Curve struct {
	Degree   int32
	Knots    []float32
	Controls []float32
}

type GR2TransformTrack struct {
	Name        string
	Position    GR2Curve
	Orientation GR2Curve
	ScaleShear  GR2Curve
}

type GR2TrackGroup struct {
	Name             string
	TransformTracks  []GR2TransformTrack
	InitialPlacement GR2Transform
}

type GR2Animation struct {
	Name              string
	Duration          float32
	TimeStep          float32
	TrackGroupIndices []int
}

type GR2File struct {
	Textures      []GR2Texture
	Materials     []GR2Material
	Skeletons     []GR2Skeleton
	VertexDatas   []GR2VertexData
	TriTopologies []GR2TriTopology
	Meshes        []GR2Mesh
	Models        []GR2Model
	TrackGroups   []GR2TrackGroup
	Animations    []GR2Animation
}

func ParseGR2(data []byte) (*GR2File, error) {
	container, err := ParseGR2Container(data)
	if err != nil {
		return nil, err
	}
	return ParseGR2File(container)
}

func ParseGR2File(container *GR2Container) (*GR2File, error) {
	nav := newGR2Nav(container.Data)
	rootType, err := container.RefOffset(container.TypeRef)
	if err != nil {
		return nil, err
	}
	rootObj, err := container.RefOffset(container.RootRef)
	if err != nil {
		return nil, err
	}
	root, err := nav.fields(rootType, rootObj)
	if err != nil {
		return nil, err
	}

	textures, textureOffsets, err := parseGR2Array(&nav, root, "Textures", parseGR2Texture)
	if err != nil {
		return nil, err
	}
	materialsRaw, materialOffsets, err := parseGR2Array(&nav, root, "Materials", parseGR2MaterialRaw)
	if err != nil {
		return nil, err
	}
	skeletons, skeletonOffsets, err := parseGR2Array(&nav, root, "Skeletons", parseGR2Skeleton)
	if err != nil {
		return nil, err
	}
	vertexDatas, vertexDataOffsets, err := parseGR2Array(&nav, root, "VertexDatas", parseGR2VertexData)
	if err != nil {
		return nil, err
	}
	triTopologies, topologyOffsets, err := parseGR2Array(&nav, root, "TriTopologies", parseGR2Topology)
	if err != nil {
		return nil, err
	}
	meshesRaw, meshOffsets, err := parseGR2Array(&nav, root, "Meshes", parseGR2MeshRaw)
	if err != nil {
		return nil, err
	}
	modelsRaw, _, err := parseGR2Array(&nav, root, "Models", parseGR2ModelRaw)
	if err != nil {
		return nil, err
	}
	trackGroups, trackGroupOffsets, err := parseGR2Array(&nav, root, "TrackGroups", parseGR2TrackGroup)
	if err != nil {
		return nil, err
	}
	animationsRaw, _, err := parseGR2Array(&nav, root, "Animations", parseGR2AnimationRaw)
	if err != nil {
		return nil, err
	}

	textureIndex := gr2OffsetIndex(textureOffsets)
	materialIndex := gr2OffsetIndex(materialOffsets)
	skeletonIndex := gr2OffsetIndex(skeletonOffsets)
	vertexDataIndex := gr2OffsetIndex(vertexDataOffsets)
	topologyIndex := gr2OffsetIndex(topologyOffsets)
	meshIndex := gr2OffsetIndex(meshOffsets)
	trackGroupIndex := gr2OffsetIndex(trackGroupOffsets)

	materials := make([]GR2Material, len(materialsRaw))
	materialRefs := make([]gr2MaterialRefs, len(materialsRaw))
	for i, raw := range materialsRaw {
		materials[i] = raw.material
		materialRefs[i] = raw.refs
		if raw.refs.texture != nil {
			if index, ok := textureIndex[*raw.refs.texture]; ok {
				materials[i].TextureIndex = intPtr(index)
			}
		}
	}
	for iter := 0; iter < 4; iter++ {
		snapshot := make([]*int, len(materials))
		for i := range materials {
			if materials[i].TextureIndex != nil {
				snapshot[i] = intPtr(*materials[i].TextureIndex)
			}
		}
		changed := false
		for i := range materials {
			if materials[i].TextureIndex != nil {
				continue
			}
			for _, off := range materialRefs[i].maps {
				mapIndex, ok := materialIndex[off]
				if ok && snapshot[mapIndex] != nil {
					materials[i].TextureIndex = intPtr(*snapshot[mapIndex])
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}

	meshes := make([]GR2Mesh, len(meshesRaw))
	for i, raw := range meshesRaw {
		mesh := raw.mesh
		if raw.refs.vertexData != nil {
			if index, ok := vertexDataIndex[*raw.refs.vertexData]; ok {
				mesh.VertexDataIndex = intPtr(index)
			}
		}
		if raw.refs.topology != nil {
			if index, ok := topologyIndex[*raw.refs.topology]; ok {
				mesh.TopologyIndex = intPtr(index)
			}
		}
		for _, off := range raw.refs.materials {
			if index, ok := materialIndex[off]; ok {
				mesh.MaterialIndices = append(mesh.MaterialIndices, index)
			}
		}
		meshes[i] = mesh
	}

	models := make([]GR2Model, len(modelsRaw))
	for i, raw := range modelsRaw {
		model := raw.model
		if raw.refs.skeleton != nil {
			if index, ok := skeletonIndex[*raw.refs.skeleton]; ok {
				model.SkeletonIndex = intPtr(index)
			}
		}
		for _, off := range raw.refs.meshes {
			if index, ok := meshIndex[off]; ok {
				model.MeshIndices = append(model.MeshIndices, index)
			}
		}
		models[i] = model
	}

	animations := make([]GR2Animation, len(animationsRaw))
	for i, raw := range animationsRaw {
		anim := raw.animation
		for _, off := range raw.trackGroups {
			if index, ok := trackGroupIndex[off]; ok {
				anim.TrackGroupIndices = append(anim.TrackGroupIndices, index)
			}
		}
		animations[i] = anim
	}

	return &GR2File{
		Textures:      textures,
		Materials:     materials,
		Skeletons:     skeletons,
		VertexDatas:   vertexDatas,
		TriTopologies: triTopologies,
		Meshes:        meshes,
		Models:        models,
		TrackGroups:   trackGroups,
		Animations:    animations,
	}, nil
}

func intPtr(v int) *int {
	return &v
}

func gr2OffsetIndex(offsets []int) map[int]int {
	out := make(map[int]int, len(offsets))
	for i, off := range offsets {
		out[off] = i
	}
	return out
}

func parseGR2Array[T any](nav *gr2Nav, root []gr2Field, name string, parse func(*gr2Nav, int, int) (T, error)) ([]T, []int, error) {
	field := gr2FindField(root, name)
	if field == nil {
		return nil, nil, nil
	}
	count, ptr, err := nav.array(field.slot)
	if err != nil {
		return nil, nil, err
	}
	out := make([]T, 0, count)
	offsets := make([]int, 0, count)
	for i := 0; i < count; i++ {
		objPtr, err := nav.u32(ptr + i*4)
		if err != nil {
			return nil, nil, err
		}
		if objPtr == 0 {
			continue
		}
		value, err := parse(nav, field.refOff, int(objPtr))
		if err != nil {
			return nil, nil, err
		}
		offsets = append(offsets, int(objPtr))
		out = append(out, value)
	}
	return out, offsets, nil
}

func parseGR2Texture(nav *gr2Nav, typeOff, obj int) (GR2Texture, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return GR2Texture{}, err
	}
	var pixels []byte
	if images := gr2FindField(fields, "Images"); images != nil {
		imgCount, imgPtr, err := nav.array(images.slot)
		if err == nil && imgCount > 0 && imgPtr != 0 {
			imgFields, err := nav.fields(images.refOff, imgPtr)
			if err == nil {
				if mips := gr2FindField(imgFields, "MIPLevels"); mips != nil {
					mipCount, mipPtr, err := nav.array(mips.slot)
					if err == nil && mipCount > 0 && mipPtr != 0 {
						mipFields, err := nav.fields(mips.refOff, mipPtr)
						if err == nil {
							if px := gr2FindField(mipFields, "Pixels"); px != nil {
								pxCount, pxPtr, err := nav.array(px.slot)
								if err == nil && pxPtr != 0 && pxPtr+pxCount <= len(nav.data) {
									pixels = append([]byte(nil), nav.data[pxPtr:pxPtr+pxCount]...)
								}
							}
						}
					}
				}
			}
		}
	}
	var bytesPerPixel int32
	var componentBits [4]int32
	if layout := gr2FindField(fields, "Layout"); layout != nil {
		layoutFields, err := nav.fields(layout.refOff, layout.slot)
		if err == nil {
			bytesPerPixel = gr2I32Field(nav, layoutFields, "BytesPerPixel")
			if bits := gr2FindField(layoutFields, "BitsForComponent"); bits != nil {
				for i := range componentBits {
					componentBits[i], _ = nav.i32(bits.slot + i*4)
				}
			}
		}
	}
	return GR2Texture{
		FromFileName:  gr2StringField(nav, fields, "FromFileName"),
		Width:         gr2I32Field(nav, fields, "Width"),
		Height:        gr2I32Field(nav, fields, "Height"),
		Encoding:      gr2I32Field(nav, fields, "Encoding"),
		SubFormat:     gr2I32Field(nav, fields, "SubFormat"),
		BytesPerPixel: bytesPerPixel,
		ComponentBits: componentBits,
		Pixels:        pixels,
	}, nil
}

type gr2MaterialRefs struct {
	texture *int
	maps    []int
}

type gr2MaterialRaw struct {
	material GR2Material
	refs     gr2MaterialRefs
}

func parseGR2MaterialRaw(nav *gr2Nav, typeOff, obj int) (gr2MaterialRaw, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return gr2MaterialRaw{}, err
	}
	refPtr := func(name string) *int {
		field := gr2FindField(fields, name)
		if field == nil {
			return nil
		}
		ptr, err := nav.u32(field.slot)
		if err != nil || ptr == 0 {
			return nil
		}
		return intPtr(int(ptr))
	}
	var maps []int
	if mapsField := gr2FindField(fields, "Maps"); mapsField != nil {
		count, ptr, err := nav.array(mapsField.slot)
		if err == nil {
			stride, err := nav.typeSize(mapsField.refOff)
			if err == nil {
				for i := 0; i < count; i++ {
					entry, err := nav.fields(mapsField.refOff, ptr+i*stride)
					if err != nil {
						return gr2MaterialRaw{}, err
					}
					if mapField := gr2FindField(entry, "Map"); mapField != nil {
						rawPtr, err := nav.u32(mapField.slot)
						if err != nil {
							return gr2MaterialRaw{}, err
						}
						if rawPtr != 0 {
							maps = append(maps, int(rawPtr))
						}
					}
				}
			}
		}
	}
	return gr2MaterialRaw{
		material: GR2Material{Name: gr2StringField(nav, fields, "Name")},
		refs:     gr2MaterialRefs{texture: refPtr("Texture"), maps: maps},
	}, nil
}

func parseGR2Skeleton(nav *gr2Nav, typeOff, obj int) (GR2Skeleton, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return GR2Skeleton{}, err
	}
	var bones []GR2Bone
	if bonesField := gr2FindField(fields, "Bones"); bonesField != nil {
		count, ptr, err := nav.array(bonesField.slot)
		if err != nil {
			return GR2Skeleton{}, err
		}
		stride, err := nav.typeSize(bonesField.refOff)
		if err != nil {
			return GR2Skeleton{}, err
		}
		for i := 0; i < count; i++ {
			boneFields, err := nav.fields(bonesField.refOff, ptr+i*stride)
			if err != nil {
				return GR2Skeleton{}, err
			}
			transform := GR2IdentityTransform
			if t := gr2FindField(boneFields, "Transform"); t != nil {
				transform, err = readGR2Transform(nav, t.slot)
				if err != nil {
					return GR2Skeleton{}, err
				}
			}
			var inverseWorld [16]float32
			if iw := gr2FindField(boneFields, "InverseWorldTransform"); iw != nil {
				for j := range inverseWorld {
					inverseWorld[j], err = nav.f32(iw.slot + j*4)
					if err != nil {
						return GR2Skeleton{}, err
					}
				}
			}
			bones = append(bones, GR2Bone{
				Name:         gr2StringField(nav, boneFields, "Name"),
				ParentIndex:  gr2I32Field(nav, boneFields, "ParentIndex"),
				Transform:    transform,
				InverseWorld: inverseWorld,
			})
		}
	}
	return GR2Skeleton{Name: gr2StringField(nav, fields, "Name"), Bones: bones}, nil
}

func parseGR2VertexData(nav *gr2Nav, typeOff, obj int) (GR2VertexData, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return GR2VertexData{}, err
	}
	var vertices []GR2Vertex
	if vf := gr2FindField(fields, "Vertices"); vf != nil {
		vtype, count, ptr, err := nav.variantArray(vf.slot)
		if err != nil {
			return GR2VertexData{}, err
		}
		if vtype != 0 && ptr != 0 {
			stride, err := nav.typeSize(vtype)
			if err != nil {
				return GR2VertexData{}, err
			}
			layout, err := nav.fields(vtype, 0)
			if err != nil {
				return GR2VertexData{}, err
			}
			comp := func(name string) *int {
				if field := gr2FindField(layout, name); field != nil {
					return intPtr(field.slot)
				}
				return nil
			}
			pos := comp("Position")
			normal := comp("Normal")
			uv := comp("TextureCoordinates0")
			weights := comp("BoneWeights")
			indices := comp("BoneIndices")
			for i := 0; i < count; i++ {
				base := ptr + i*stride
				read3 := func(off *int) ([3]float32, error) {
					var out [3]float32
					if off == nil {
						return out, nil
					}
					for j := range out {
						value, err := nav.f32(base + *off + j*4)
						if err != nil {
							return out, err
						}
						out[j] = value
					}
					return out, nil
				}
				position, err := read3(pos)
				if err != nil {
					return GR2VertexData{}, err
				}
				nrm, err := read3(normal)
				if err != nil {
					return GR2VertexData{}, err
				}
				var uv2 [2]float32
				if uv != nil {
					uv2[0], err = nav.f32(base + *uv)
					if err != nil {
						return GR2VertexData{}, err
					}
					uv2[1], err = nav.f32(base + *uv + 4)
					if err != nil {
						return GR2VertexData{}, err
					}
				}
				read4 := func(off *int) [4]byte {
					var out [4]byte
					if off == nil {
						return out
					}
					for j := range out {
						if base+*off+j < len(nav.data) {
							out[j] = nav.data[base+*off+j]
						}
					}
					return out
				}
				vertices = append(vertices, GR2Vertex{
					Position:    position,
					Normal:      nrm,
					UV:          uv2,
					BoneWeights: read4(weights),
					BoneIndices: read4(indices),
				})
			}
		}
	}
	return GR2VertexData{Vertices: vertices}, nil
}

func parseGR2Topology(nav *gr2Nav, typeOff, obj int) (GR2TriTopology, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return GR2TriTopology{}, err
	}
	var groups []GR2TriGroup
	if gf := gr2FindField(fields, "Groups"); gf != nil {
		count, ptr, err := nav.array(gf.slot)
		if err != nil {
			return GR2TriTopology{}, err
		}
		stride, err := nav.typeSize(gf.refOff)
		if err != nil {
			return GR2TriTopology{}, err
		}
		for i := 0; i < count; i++ {
			groupFields, err := nav.fields(gf.refOff, ptr+i*stride)
			if err != nil {
				return GR2TriTopology{}, err
			}
			groups = append(groups, GR2TriGroup{
				MaterialIndex: gr2I32Field(nav, groupFields, "MaterialIndex"),
				TriFirst:      gr2I32Field(nav, groupFields, "TriFirst"),
				TriCount:      gr2I32Field(nav, groupFields, "TriCount"),
			})
		}
	}
	indices, err := parseGR2TopologyIndices(nav, fields, "Indices", 4)
	if err != nil {
		return GR2TriTopology{}, err
	}
	if len(indices) == 0 {
		indices, err = parseGR2TopologyIndices(nav, fields, "Indices16", 2)
		if err != nil {
			return GR2TriTopology{}, err
		}
	}
	return GR2TriTopology{Groups: groups, Indices: indices}, nil
}

func parseGR2TopologyIndices(nav *gr2Nav, fields []gr2Field, name string, bytesPerIndex int) ([]uint32, error) {
	field := gr2FindField(fields, name)
	if field == nil {
		return nil, nil
	}
	count, ptr, err := nav.array(field.slot)
	if err != nil || ptr == 0 {
		return nil, err
	}
	indices := make([]uint32, 0, count)
	for i := 0; i < count; i++ {
		switch bytesPerIndex {
		case 4:
			value, err := nav.u32(ptr + i*4)
			if err != nil {
				return nil, err
			}
			indices = append(indices, value)
		case 2:
			off := ptr + i*2
			if off+2 > len(nav.data) {
				return nil, fmt.Errorf("gr2: index buffer truncated")
			}
			indices = append(indices, uint32(binary.LittleEndian.Uint16(nav.data[off:off+2])))
		}
	}
	return indices, nil
}

type gr2MeshRefs struct {
	vertexData *int
	topology   *int
	materials  []int
}

type gr2MeshRaw struct {
	mesh GR2Mesh
	refs gr2MeshRefs
}

func parseGR2MeshRaw(nav *gr2Nav, typeOff, obj int) (gr2MeshRaw, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return gr2MeshRaw{}, err
	}
	refPtr := func(name string) *int {
		field := gr2FindField(fields, name)
		if field == nil {
			return nil
		}
		ptr, err := nav.u32(field.slot)
		if err != nil || ptr == 0 {
			return nil
		}
		return intPtr(int(ptr))
	}
	var materials []int
	if mb := gr2FindField(fields, "MaterialBindings"); mb != nil {
		count, ptr, err := nav.array(mb.slot)
		if err != nil {
			return gr2MeshRaw{}, err
		}
		stride, err := nav.typeSize(mb.refOff)
		if err != nil {
			return gr2MeshRaw{}, err
		}
		for i := 0; i < count; i++ {
			bindingFields, err := nav.fields(mb.refOff, ptr+i*stride)
			if err != nil {
				return gr2MeshRaw{}, err
			}
			if mf := gr2FindField(bindingFields, "Material"); mf != nil {
				ptr, err := nav.u32(mf.slot)
				if err != nil {
					return gr2MeshRaw{}, err
				}
				if ptr != 0 {
					materials = append(materials, int(ptr))
				}
			}
		}
	}
	var boneBindings []string
	if bb := gr2FindField(fields, "BoneBindings"); bb != nil {
		count, ptr, err := nav.array(bb.slot)
		if err != nil {
			return gr2MeshRaw{}, err
		}
		stride, err := nav.typeSize(bb.refOff)
		if err != nil {
			return gr2MeshRaw{}, err
		}
		for i := 0; i < count; i++ {
			bindingFields, err := nav.fields(bb.refOff, ptr+i*stride)
			if err != nil {
				return gr2MeshRaw{}, err
			}
			boneBindings = append(boneBindings, gr2StringField(nav, bindingFields, "BoneName"))
		}
	}
	return gr2MeshRaw{
		mesh: GR2Mesh{
			Name:         gr2StringField(nav, fields, "Name"),
			BoneBindings: boneBindings,
		},
		refs: gr2MeshRefs{
			vertexData: refPtr("PrimaryVertexData"),
			topology:   refPtr("PrimaryTopology"),
			materials:  materials,
		},
	}, nil
}

type gr2ModelRefs struct {
	skeleton *int
	meshes   []int
}

type gr2ModelRaw struct {
	model GR2Model
	refs  gr2ModelRefs
}

func parseGR2ModelRaw(nav *gr2Nav, typeOff, obj int) (gr2ModelRaw, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return gr2ModelRaw{}, err
	}
	var skeleton *int
	if sf := gr2FindField(fields, "Skeleton"); sf != nil {
		ptr, err := nav.u32(sf.slot)
		if err != nil {
			return gr2ModelRaw{}, err
		}
		if ptr != 0 {
			skeleton = intPtr(int(ptr))
		}
	}
	var meshes []int
	if mb := gr2FindField(fields, "MeshBindings"); mb != nil {
		count, ptr, err := nav.array(mb.slot)
		if err != nil {
			return gr2ModelRaw{}, err
		}
		stride, err := nav.typeSize(mb.refOff)
		if err != nil {
			return gr2ModelRaw{}, err
		}
		for i := 0; i < count; i++ {
			bindingFields, err := nav.fields(mb.refOff, ptr+i*stride)
			if err != nil {
				return gr2ModelRaw{}, err
			}
			if mf := gr2FindField(bindingFields, "Mesh"); mf != nil {
				ptr, err := nav.u32(mf.slot)
				if err != nil {
					return gr2ModelRaw{}, err
				}
				if ptr != 0 {
					meshes = append(meshes, int(ptr))
				}
			}
		}
	}
	initialPlacement := GR2IdentityTransform
	if ip := gr2FindField(fields, "InitialPlacement"); ip != nil {
		initialPlacement, err = readGR2Transform(nav, ip.slot)
		if err != nil {
			return gr2ModelRaw{}, err
		}
	}
	return gr2ModelRaw{
		model: GR2Model{
			Name:             gr2StringField(nav, fields, "Name"),
			InitialPlacement: initialPlacement,
		},
		refs: gr2ModelRefs{skeleton: skeleton, meshes: meshes},
	}, nil
}

func parseGR2Curve(nav *gr2Nav, typeOff, slot int) (GR2Curve, error) {
	fields, err := nav.fields(typeOff, slot)
	if err != nil {
		return GR2Curve{}, err
	}
	readReals := func(name string) ([]float32, error) {
		field := gr2FindField(fields, name)
		if field == nil {
			return nil, nil
		}
		count, ptr, err := nav.array(field.slot)
		if err != nil || ptr == 0 {
			return nil, err
		}
		out := make([]float32, 0, count)
		for i := 0; i < count; i++ {
			value, err := nav.f32(ptr + i*4)
			if err != nil {
				return nil, err
			}
			out = append(out, value)
		}
		return out, nil
	}
	knots, err := readReals("Knots")
	if err != nil {
		return GR2Curve{}, err
	}
	controls, err := readReals("Controls")
	if err != nil {
		return GR2Curve{}, err
	}
	return GR2Curve{
		Degree:   gr2I32Field(nav, fields, "Degree"),
		Knots:    knots,
		Controls: controls,
	}, nil
}

func parseGR2TrackGroup(nav *gr2Nav, typeOff, obj int) (GR2TrackGroup, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return GR2TrackGroup{}, err
	}
	var tracks []GR2TransformTrack
	if tt := gr2FindField(fields, "TransformTracks"); tt != nil {
		count, ptr, err := nav.array(tt.slot)
		if err != nil {
			return GR2TrackGroup{}, err
		}
		stride, err := nav.typeSize(tt.refOff)
		if err != nil {
			return GR2TrackGroup{}, err
		}
		for i := 0; i < count; i++ {
			trackFields, err := nav.fields(tt.refOff, ptr+i*stride)
			if err != nil {
				return GR2TrackGroup{}, err
			}
			curve := func(name string) (GR2Curve, error) {
				field := gr2FindField(trackFields, name)
				if field == nil {
					return GR2Curve{}, nil
				}
				return parseGR2Curve(nav, field.refOff, field.slot)
			}
			position, err := curve("PositionCurve")
			if err != nil {
				return GR2TrackGroup{}, err
			}
			orientation, err := curve("OrientationCurve")
			if err != nil {
				return GR2TrackGroup{}, err
			}
			scaleShear, err := curve("ScaleShearCurve")
			if err != nil {
				return GR2TrackGroup{}, err
			}
			tracks = append(tracks, GR2TransformTrack{
				Name:        gr2StringField(nav, trackFields, "Name"),
				Position:    position,
				Orientation: orientation,
				ScaleShear:  scaleShear,
			})
		}
	}
	initialPlacement := GR2IdentityTransform
	if ip := gr2FindField(fields, "InitialPlacement"); ip != nil {
		initialPlacement, err = readGR2Transform(nav, ip.slot)
		if err != nil {
			return GR2TrackGroup{}, err
		}
	}
	return GR2TrackGroup{
		Name:             gr2StringField(nav, fields, "Name"),
		TransformTracks:  tracks,
		InitialPlacement: initialPlacement,
	}, nil
}

type gr2AnimationRaw struct {
	animation   GR2Animation
	trackGroups []int
}

func parseGR2AnimationRaw(nav *gr2Nav, typeOff, obj int) (gr2AnimationRaw, error) {
	fields, err := nav.fields(typeOff, obj)
	if err != nil {
		return gr2AnimationRaw{}, err
	}
	var trackGroups []int
	if tg := gr2FindField(fields, "TrackGroups"); tg != nil {
		count, ptr, err := nav.array(tg.slot)
		if err != nil {
			return gr2AnimationRaw{}, err
		}
		for i := 0; i < count; i++ {
			rawPtr, err := nav.u32(ptr + i*4)
			if err != nil {
				return gr2AnimationRaw{}, err
			}
			if rawPtr != 0 {
				trackGroups = append(trackGroups, int(rawPtr))
			}
		}
	}
	return gr2AnimationRaw{
		animation: GR2Animation{
			Name:     gr2StringField(nav, fields, "Name"),
			Duration: gr2F32Field(nav, fields, "Duration"),
			TimeStep: gr2F32Field(nav, fields, "TimeStep"),
		},
		trackGroups: trackGroups,
	}, nil
}

func IsGR2ResourceName(name string) bool {
	name = strings.ToLower(strings.ReplaceAll(name, "/", "\\"))
	if i := strings.LastIndex(name, "\\"); i >= 0 {
		name = name[i+1:]
	}
	return strings.HasSuffix(name, ".gr2")
}

func GR2ModelResourceCandidates(resourceName string) []string {
	resourceName = strings.TrimSpace(strings.ReplaceAll(resourceName, "/", "\\"))
	if resourceName == "" {
		return nil
	}
	if i := strings.LastIndex(resourceName, "\\"); i >= 0 {
		resourceName = resourceName[i+1:]
	}
	lower := strings.ToLower(resourceName)
	return dedupeStrings([]string{
		"data\\model\\3dmob\\" + lower,
		"data/model/3dmob/" + lower,
		"model\\3dmob\\" + lower,
		"model/3dmob/" + lower,
		lower,
		resourceName,
	})
}

func GR2BoneTypeFromName(name string) (uint32, bool) {
	name = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(name, "/", "\\")))
	if i := strings.LastIndex(name, "\\"); i >= 0 {
		name = name[i+1:]
	}
	name = strings.TrimSuffix(name, ".gr2")
	underscore := strings.LastIndex(name, "_")
	if underscore < 0 || underscore+1 >= len(name) {
		return 0, false
	}
	var out uint32
	for _, ch := range name[underscore+1:] {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		out = out*10 + uint32(ch-'0')
	}
	return out, true
}

type GR2Action int

const (
	GR2ActionStand GR2Action = iota
	GR2ActionMove
	GR2ActionAttack
	GR2ActionDead
	GR2ActionDamage
)

func GR2AnimationResourceCandidates(boneType uint32, action GR2Action) []string {
	suffix := ""
	switch action {
	case GR2ActionMove:
		suffix = "move"
	case GR2ActionAttack:
		suffix = "attack"
	case GR2ActionDead:
		suffix = "dead"
	case GR2ActionDamage:
		suffix = "damage"
	default:
		return nil
	}
	name := fmt.Sprintf("%d_%s.gr2", boneType, suffix)
	return []string{
		"data\\model\\3dmob_bone\\" + name,
		"data/model/3dmob_bone/" + name,
		"model\\3dmob_bone\\" + name,
		"model/3dmob_bone/" + name,
		name,
	}
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
