/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package storekeeper

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStreamer_NewStreamer(t *testing.T) {
	s := NewStreamer()
	s.Paths = append(s.Paths, "gene/Entrez Gene/suk_entrez_gene_id_i/GID", "gene/Entrez Gene/udk_cid_s/CID")
	s.Paths = append(s.Paths, "protein/UniProtKB/suk_uniprot_id_i/PID")
	s.Items = append(s.Items, []any{1, 2, 3, 4, 5}, []any{"BAH", "BAU", "BAH", "BAN", "BAC"})
	s.Items = append(s.Items, []any{100, 200, 300})
	require.NoError(t, s.IntegrityCheck())
}

func TestStreamer_IntegrityCheck(t *testing.T) {
	s := createTestDataset(t)
	require.NoError(t, s.Add("gene/Entrez Gene/udk_fail_s/Bad", []any{20, 30}))
	require.Error(t, s.IntegrityCheck())
	s.Reset()
	require.NotNil(t, s)
	require.NoError(t, s.Add("gene/Entrez Gene/suk_entrez_gene_id_i/GID", []any{1, 2, 3, 4, 5}))
	require.NoError(t, s.IntegrityCheck())
	require.NoError(t, s.Add("gene/Entrez Gene/udk_cid_s/CID/Other stuff", []any{"BAH", "BAU", "BAH", "BAN", "BAC"}))
	require.NoError(t, s.IntegrityCheck())
	require.NoError(t, s.Add("protein/suk_uniprot_id_i/PID", []any{100, 200, 300}))
	require.NoError(t, s.IntegrityCheck())
	require.NoError(t, s.Add("protein/suk_uniprot_id_i/PID", []any{100, 200, 300, 400}))
	require.Error(t, s.IntegrityCheck())
}

func TestStreamer_Add(t *testing.T) {
	s := NewStreamer()
	require.NotNil(t, s)
	require.NoError(t, s.Add("gene/Entrez Gene/suk_entrez_gene_id_i/GID", []any{1, 2, 3, 4, 5}))
	require.NoError(t, s.Add("protein/UniProtKB/suk_uniprot_id_i/PID", []any{100, 200, 300}))
	require.NoError(t, s.IntegrityCheck())
	s.Items = append(s.Items, []any{"BAH", "BAU", "BAH", "BAN", "BAC"})
	require.NoError(t, s.Add("protein/UniProtKB/suk_uniprot_id_i/PID", []any{100, 200, 300}))
}

func TestStreamer_Reset(t *testing.T) {
	s := NewStreamer()
	require.NotNil(t, s)
	require.NoError(t, s.Add("gene/Entrez Gene/suk_entrez_gene_id_i/GID", []any{1, 2, 3, 4, 5}))
	require.NoError(t, s.Add("protein/UniProtKB/suk_uniprot_id_i/PID", []any{100, 200, 300}))
	require.NoError(t, s.IntegrityCheck())
	s.Reset()
	require.NotNil(t, s)
	require.NotNil(t, s.Items)
	require.NotNil(t, s.Paths)
	require.Equal(t, 0, len(s.Items))
	require.Equal(t, 0, len(s.Paths))
	require.Empty(t, s.Msg)
	require.Nil(t, s.Err)
}

func TestStreamer_String(t *testing.T) {
	s := createTestDataset(t)
	kds, err := KeyNamespace(s.Paths[0])
	require.NoError(t, err)
	require.Equal(t, "1\tBAH", s.String(kds, 0, "\t"))
	require.Equal(t, "2\tBAU", s.String(kds, 1, "\t"))
	require.Equal(t, "3\tBAH", s.String(kds, 2, "\t"))
	require.Equal(t, "4|BAN", s.String(kds, 3, "|"))
	require.Equal(t, "5;BAC", s.String(kds, 4, ";"))
	require.NotEqual(t, "5;BAC", s.String(kds, 4, "\t"))
}

func TestStreamer_BufEncode(t *testing.T) {
	var network bytes.Buffer
	s := createTestDataset(t)
	require.NoError(t, s.BufEncode(&network))
}

func TestStreamer_BufDecode(t *testing.T) {
	var network bytes.Buffer
	s := createTestDataset(t)
	require.NoError(t, s.BufEncode(&network))
	r, err := BufDecode(&network)
	require.NoError(t, err)
	require.Equal(t, s.Paths, r.Paths)
	require.Equal(t, s.Items, r.Items)
}

func TestStreamer_CmdEncode_CmdDecode(t *testing.T) {
	// Start a server to receive the encoded data
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	listener, ret := net.Listen("tcp", "127.0.0.1:11404")
	require.NoError(t, ret)
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:11404")
	require.NoError(t, err)
	defer func(conn net.Conn) {
		err = conn.Close()
		require.NoError(t, err)
	}(conn)
	// Create the test data
	s := createTestDataset(t)
	// Encode the Gob package, specify the request command for an end-point server, and send it.
	require.NoError(t, s.CmdEncode(conn, "create/gene/document"))
	// Listen and receive from the network the streamed data.
	conn, err = listener.Accept()
	require.NoError(t, err)
	// Parse the end-point cmd and decode the Gob package from the streamed data.
	r, cmd, fer := CmdDecode(conn)
	require.NoError(t, fer)
	require.NotNil(t, r)
	require.Equal(t, "create/gene/document", cmd)
	require.Equal(t, 3, len(r.Paths))
	require.Equal(t, s.Paths, r.Paths)
	require.Equal(t, s.Items, r.Items)
}

func TestStreamer_NetEncode_NetDecode(t *testing.T) {
	// Start a server to receive the encoded data
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	listener, ret := net.Listen("tcp", "127.0.0.1:11405")
	require.NoError(t, ret)
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:11405")
	require.NoError(t, err)
	defer func(conn net.Conn) {
		err = conn.Close()
		require.NoError(t, err)
	}(conn)
	// Create the test data
	s := createTestDataset(t)
	// Encode the Gob package and send it.
	require.NoError(t, s.NetEncode(conn))
	// Listen and receive from the network the streamed data.
	conn, err = listener.Accept()
	require.NoError(t, err)
	// Decode the Gob package from the streamed data.
	r, fer := NetDecode(conn)
	require.NoError(t, fer)
	require.NotNil(t, r)
	require.Equal(t, 3, len(r.Paths))
	require.Equal(t, s.Paths, r.Paths)
	require.Equal(t, s.Items, r.Items)
}

func createTestDataset(t *testing.T) *Streamer {
	s := NewStreamer()
	require.NotNil(t, s)
	require.NoError(t, s.Add("gene/Entrez Gene/suk_entrez_gene_id_i/GID", []any{1, 2, 3, 4, 5}))
	require.NoError(t, s.Add("protein/UniProtKB/suk_uniprot_id_i/PID", []any{100, 200, 300}))
	require.NoError(t, s.Add("gene/Entrez Gene/udk_cid_s/CID", []any{"BAH", "BAU", "BAH", "BAN", "BAC"}))
	require.NoError(t, s.IntegrityCheck())
	return s
}
