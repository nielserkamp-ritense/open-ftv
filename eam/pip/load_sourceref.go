package pip

import (
	"fmt"
	"os"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// fileSourceRef builds a "file" source reference for a file-store path. Unlike a
// pulled value (which references a logged WARC exchange), a file-store value is
// referenced by the file that holds it, versioned by the file's size+modtime as a
// cheap content-change signal. It returns nil when the PIP tracks no source
// references, so file loading stays a no-op cost when refs are disabled.
//
// The path is carried in the WARCFile field (the "file that holds the value"),
// analogous to the WARC exchange for pulled values; Version carries the mtime tag.
func (p *pip) fileSourceRef(path string) *SourceRef {
	if p.sourceRefs == nil {
		return nil
	}
	version := path
	if fi, err := os.Stat(path); err == nil {
		version = fmt.Sprintf("%d-%d", fi.Size(), fi.ModTime().UnixNano())
	}
	return &SourceRef{Kind: "file", WARCFile: path, Version: version}
}

// recordFileRef registers a single file-store source reference under key. It writes
// straight to the source-ref store (rather than via RecordSourceRef) so that bulk
// file loading does not emit a debug line per value.
func (p *pip) recordFileRef(key string, ref *SourceRef) {
	if ref == nil || key == "" {
		return
	}
	p.sourceRefs.put(SourceRef{Kind: ref.Kind, Key: key, WARCFile: ref.WARCFile, Version: ref.Version})
}

// recordEntityFileRef registers file-store source references for an entity: one
// under the entity UID and one per attribute under "<uid>/<attr>", matching the key
// shape a PDP request delivers so the ADL level-3 hook can resolve it.
func (p *pip) recordEntityFileRef(e models.Entity, ref *SourceRef) {
	if ref == nil || e == nil {
		return
	}
	uid := e.UID()
	p.recordFileRef(uid, ref)
	if attrs := e.Attributes(); attrs != nil {
		attrs.IterateAttributes(func(a models.Attribute) {
			p.recordFileRef(uid+"/"+a.Key(), ref)
		})
	}
}
