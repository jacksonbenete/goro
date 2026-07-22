package res

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type GR2ModelVertex struct {
	Position    [3]float32
	Normal      [3]float32
	UV          [2]float32
	BoneIndices [4]byte
	BoneWeights [4]byte
}

type GR2GeometryBatch struct {
	TextureIndex int
	StartIndex   uint32
	IndexCount   uint32
}

type GR2Geometry struct {
	Vertices  []GR2ModelVertex
	Indices   []uint32
	Batches   []GR2GeometryBatch
	BoneCount int
	Center    [3]float32
	Size      [3]float32
}

func BuildGR2Geometry(file *GR2File, modelIndex int) (*GR2Geometry, error) {
	if file == nil {
		return nil, fmt.Errorf("gr2 geometry: nil file")
	}
	if modelIndex < 0 || modelIndex >= len(file.Models) {
		return nil, fmt.Errorf("gr2 geometry: missing model %d", modelIndex)
	}
	model := file.Models[modelIndex]
	if model.SkeletonIndex == nil || *model.SkeletonIndex < 0 || *model.SkeletonIndex >= len(file.Skeletons) {
		return nil, fmt.Errorf("gr2 geometry: missing skeleton")
	}
	skeleton := file.Skeletons[*model.SkeletonIndex]
	boneByName := make(map[string]byte, len(skeleton.Bones))
	for i, bone := range skeleton.Bones {
		boneByName[bone.Name] = byte(i)
	}

	var vertices []GR2ModelVertex
	perTexture := make(map[int][]uint32)

	for _, meshIndex := range model.MeshIndices {
		if meshIndex < 0 || meshIndex >= len(file.Meshes) {
			continue
		}
		mesh := file.Meshes[meshIndex]
		if mesh.VertexDataIndex == nil || *mesh.VertexDataIndex < 0 || *mesh.VertexDataIndex >= len(file.VertexDatas) {
			continue
		}
		if mesh.TopologyIndex == nil || *mesh.TopologyIndex < 0 || *mesh.TopologyIndex >= len(file.TriTopologies) {
			continue
		}
		vertexData := file.VertexDatas[*mesh.VertexDataIndex]
		topology := file.TriTopologies[*mesh.TopologyIndex]

		bindingToBone := make([]byte, len(mesh.BoneBindings))
		for i, name := range mesh.BoneBindings {
			bindingToBone[i] = boneByName[name]
		}
		remap := func(slot byte) byte {
			if int(slot) >= 0 && int(slot) < len(bindingToBone) {
				return bindingToBone[slot]
			}
			return 0
		}
		groupTexture := func(group GR2TriGroup) int {
			materialSlot := int(group.MaterialIndex)
			if materialSlot < 0 {
				materialSlot = 0
			}
			if materialSlot >= len(mesh.MaterialIndices) {
				return 0
			}
			materialIndex := mesh.MaterialIndices[materialSlot]
			if materialIndex < 0 || materialIndex >= len(file.Materials) {
				return 0
			}
			textureIndex := file.Materials[materialIndex].TextureIndex
			if textureIndex == nil {
				return 0
			}
			return *textureIndex
		}

		emblemMesh := len(topology.Groups) > 0
		for _, group := range topology.Groups {
			textureIndex := groupTexture(group)
			if textureIndex < 0 || textureIndex >= len(file.Textures) || !strings.Contains(strings.ToLower(file.Textures[textureIndex].FromFileName), "emblem") {
				emblemMesh = false
				break
			}
		}
		xShift := float32(0)
		if emblemMesh && len(vertexData.Vertices) > 0 {
			minX, maxX := float32(math.MaxFloat32), float32(-math.MaxFloat32)
			for _, vertex := range vertexData.Vertices {
				if vertex.Position[0] < minX {
					minX = vertex.Position[0]
				}
				if vertex.Position[0] > maxX {
					maxX = vertex.Position[0]
				}
			}
			xShift = (minX + maxX) * 0.5
		}

		base := uint32(len(vertices))
		for _, vertex := range vertexData.Vertices {
			slots, weights := vertex.BoneIndices, vertex.BoneWeights
			if weights == [4]byte{} {
				slots = [4]byte{}
				weights = [4]byte{255, 0, 0, 0}
			}
			vertices = append(vertices, GR2ModelVertex{
				Position:    [3]float32{vertex.Position[0] - xShift, vertex.Position[1], vertex.Position[2]},
				Normal:      vertex.Normal,
				UV:          vertex.UV,
				BoneIndices: [4]byte{remap(slots[0]), remap(slots[1]), remap(slots[2]), remap(slots[3])},
				BoneWeights: weights,
			})
		}

		for _, group := range topology.Groups {
			textureIndex := groupTexture(group)
			first := int(group.TriFirst)
			if first < 0 {
				first = 0
			}
			count := int(group.TriCount)
			if count < 0 {
				count = 0
			}
			first *= 3
			count *= 3
			if first < 0 || first+count > len(topology.Indices) {
				continue
			}
			for _, index := range topology.Indices[first : first+count] {
				perTexture[textureIndex] = append(perTexture[textureIndex], base+index)
			}
		}
	}

	if len(vertices) == 0 {
		return nil, fmt.Errorf("gr2 geometry: no vertices")
	}
	indices := make([]uint32, 0)
	batches := make([]GR2GeometryBatch, 0, len(perTexture))
	keys := make([]int, 0, len(perTexture))
	for key := range perTexture {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	for _, key := range keys {
		idxs := perTexture[key]
		batches = append(batches, GR2GeometryBatch{
			TextureIndex: key,
			StartIndex:   uint32(len(indices)),
			IndexCount:   uint32(len(idxs)),
		})
		indices = append(indices, idxs...)
	}

	minPos := [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	maxPos := [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	for _, vertex := range vertices {
		for i := 0; i < 3; i++ {
			if vertex.Position[i] < minPos[i] {
				minPos[i] = vertex.Position[i]
			}
			if vertex.Position[i] > maxPos[i] {
				maxPos[i] = vertex.Position[i]
			}
		}
	}
	var center, size [3]float32
	for i := 0; i < 3; i++ {
		center[i] = (minPos[i] + maxPos[i]) * 0.5
		size[i] = maxPos[i] - minPos[i]
	}

	return &GR2Geometry{
		Vertices:  vertices,
		Indices:   indices,
		Batches:   batches,
		BoneCount: len(skeleton.Bones),
		Center:    center,
		Size:      size,
	}, nil
}
