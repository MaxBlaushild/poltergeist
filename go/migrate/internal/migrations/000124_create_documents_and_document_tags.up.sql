-- Create document_tags table
CREATE TABLE document_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    text VARCHAR NOT NULL
);

CREATE INDEX idx_document_tags_text ON document_tags(text);

-- Create documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    title VARCHAR NOT NULL,
    provider VARCHAR NOT NULL,
    user_id UUID NOT NULL,
    link TEXT,
    content TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_documents_user_id ON documents(user_id);
CREATE INDEX idx_documents_provider ON documents(provider);

-- Create join table for documents and document_tags
-- GORM many2many association will use document_id and document_tag_id as foreign keys
CREATE TABLE document_document_tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    document_id UUID NOT NULL,
    document_tag_id UUID NOT NULL,
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (document_tag_id) REFERENCES document_tags(id) ON DELETE CASCADE,
    UNIQUE(document_id, document_tag_id)
);

CREATE INDEX idx_document_document_tags_document_id ON document_document_tags(document_id);
CREATE INDEX idx_document_document_tags_document_tag_id ON document_document_tags(document_tag_id);

