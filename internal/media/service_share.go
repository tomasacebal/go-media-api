package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateShare crea un link corto para archivo o envio.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - includeAll: permite admin global.
//   - targetType: file o transfer.
//   - targetID: id del recurso.
//   - expiresDays: dias de vigencia.
//
// Returns:
//   - Share creado o error.
func (s *Service) CreateShare(ctx context.Context, ownerID string, includeAll bool, targetType string, targetID string, expiresDays int, neverExpires bool) (Share, error) {
	targetType = strings.TrimSpace(targetType)
	targetID = strings.TrimSpace(targetID)
	if err := s.validateShareTarget(ctx, ownerID, includeAll, targetType, targetID); err != nil {
		return Share{}, err
	}
	code, err := s.generateShareCode(ctx)
	if err != nil {
		return Share{}, appError("share_code_failed", "No se pudo generar el link", err)
	}
	now := time.Now().UTC()
	expiresAt := defaultExpiry(resolveExpiryDays(expiresDays, s.shareTTLDays))
	share := Share{
		ID:           uuid.NewString(),
		OwnerID:      ownerID,
		Code:         code,
		URL:          s.shareURL(code),
		TargetType:   targetType,
		TargetID:     targetID,
		ExpiresAt:    &expiresAt,
		NeverExpires: neverExpires,
		CreatedAt:    now,
	}
	if neverExpires {
		share.ExpiresAt = nil
	}
	if targetType == ShareTargetFile {
		file, err := s.repo.FindFileByID(ctx, targetID)
		if err == nil {
			share.SharedFiles = []string{file.OriginalName}
		}
	} else if targetType == ShareTargetTransfer {
		files, err := s.repo.ListFilesByTransferID(ctx, targetID)
		if err == nil {
			fileNames := make([]string, 0, len(files))
			for _, f := range files {
				fileNames = append(fileNames, f.OriginalName)
			}
			share.SharedFiles = fileNames
		}
	}
	if err := s.repo.CreateShare(ctx, share); err != nil {
		return Share{}, appError("share_create_failed", "No se pudo crear el link", err)
	}
	return share, nil
}

// ListShares devuelve shares visibles.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//   - limit: cantidad maxima.
//
// Returns:
//   - Shares o error.
func (s *Service) ListShares(ctx context.Context, ownerID string, includeAll bool, limit int) ([]Share, error) {
	shares, err := s.repo.ListShares(ctx, ShareFilter{OwnerID: ownerID, IncludeAll: includeAll, Limit: limit})
	if err != nil {
		return nil, appError("share_list_failed", "No se pudieron listar links", err)
	}
	for i := range shares {
		shares[i].URL = s.shareURL(shares[i].Code)
		if shares[i].TargetType == ShareTargetFile {
			file, err := s.repo.FindFileByID(ctx, shares[i].TargetID)
			if err == nil {
				shares[i].SharedFiles = []string{file.OriginalName}
			}
		} else if shares[i].TargetType == ShareTargetTransfer {
			files, err := s.repo.ListFilesByTransferID(ctx, shares[i].TargetID)
			if err == nil {
				fileNames := make([]string, 0, len(files))
				for _, f := range files {
					fileNames = append(fileNames, f.OriginalName)
				}
				shares[i].SharedFiles = fileNames
			}
		}
	}
	return shares, nil
}

// RevokeShare revoca un link.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador del share.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//
// Returns:
//   - Error si no pudo revocar.
func (s *Service) RevokeShare(ctx context.Context, id string, ownerID string, includeAll bool) error {
	err := s.repo.RevokeShare(ctx, id, ownerID, includeAll, time.Now().UTC())
	if err != nil {
		return appError("share_revoke_failed", "No se pudo revocar el link", err)
	}
	return nil
}

// ResolveShare devuelve detalle publico de un codigo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - code: codigo corto.
//
// Returns:
//   - Detalle publico o error.
func (s *Service) ResolveShare(ctx context.Context, code string) (ShareDetail, error) {
	share, err := s.repo.FindShareByCode(ctx, code)
	if err != nil {
		return ShareDetail{}, appError("share_not_found", "Link no encontrado", ErrNotFound)
	}
	if !share.NeverExpires && share.ExpiresAt != nil && time.Now().UTC().After(*share.ExpiresAt) {
		return ShareDetail{}, appError("share_expired", "Link vencido", ErrShareExpired)
	}
	share.URL = s.shareURL(share.Code)
	return s.detailForShare(ctx, share)
}

// OpenShare prepara descarga publica.
//
// Args:
//   - ctx: contexto de la operacion.
//   - code: codigo corto.
//
// Returns:
//   - Detalle publico o error.
func (s *Service) OpenShare(ctx context.Context, code string) (ShareDetail, error) {
	return s.ResolveShare(ctx, code)
}

func (s *Service) detailForShare(ctx context.Context, share Share) (ShareDetail, error) {
	switch share.TargetType {
	case ShareTargetFile:
		file, err := s.repo.FindFileByID(ctx, share.TargetID)
		if err != nil {
			return ShareDetail{}, appError("share_not_found", "Link no encontrado", ErrNotFound)
		}
		return ShareDetail{Share: share, Files: []File{file}}, nil
	case ShareTargetTransfer:
		transfer, err := s.repo.FindTransferByID(ctx, share.TargetID)
		if err != nil {
			return ShareDetail{}, appError("share_not_found", "Link no encontrado", ErrNotFound)
		}
		files, err := s.repo.ListFilesByTransferID(ctx, transfer.ID)
		if err != nil || len(files) == 0 {
			return ShareDetail{}, appError("share_not_found", "Link no encontrado", ErrNotFound)
		}
		return ShareDetail{Share: share, Transfer: &transfer, Files: files}, nil
	default:
		return ShareDetail{}, appError("share_invalid", "Link invalido", ErrInvalidShare)
	}
}

func (s *Service) validateShareTarget(ctx context.Context, ownerID string, includeAll bool, targetType string, targetID string) error {
	switch targetType {
	case ShareTargetFile:
		_, err := s.GetFile(ctx, targetID, ownerID, includeAll)
		return err
	case ShareTargetTransfer:
		_, err := s.GetTransfer(ctx, targetID, ownerID, includeAll)
		return err
	default:
		return appError("share_invalid_target", "Destino de link invalido", ErrInvalidShare)
	}
}

func (s *Service) generateShareCode(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 8; attempt++ {
		code, err := randomBase62(defaultShareCodeLength)
		if err != nil {
			return "", err
		}
		exists, err := s.repo.ShareCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("no se pudo generar codigo unico")
}
