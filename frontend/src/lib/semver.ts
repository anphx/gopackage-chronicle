// Classifies Go module versions by their shape. Mirrors versionKindSQL in
// backend/internal/repository/stats.go so the UI and the API agree.

import type { ReleaseKindName } from './api/client';

export type ReleaseKind = ReleaseKindName;

export const RELEASE_KINDS: ReleaseKind[] = ['major', 'minor', 'patch', 'prerelease'];

export const KIND_LABEL: Record<ReleaseKind, string> = {
	major: 'Major',
	minor: 'Minor',
	patch: 'Patch',
	prerelease: 'Pre-release'
};

/**
 * Returns the kind of change a version represents, judged by its shape alone:
 * a pre-release part (including pseudo-versions) makes it a pre-release, vX.0.0 is a major,
 * vX.Y.0 a minor and anything else a patch. Build metadata such as "+incompatible" is ignored.
 */
export function releaseKind(version: string): ReleaseKind {
	const core = version.split('+', 1)[0];
	if (core.includes('-')) return 'prerelease';
	if (/^v\d+\.0\.0$/.test(core)) return 'major';
	if (/^v\d+\.\d+\.0$/.test(core)) return 'minor';
	return 'patch';
}

/** Drops the leading "v" for compact labels: "v3.0.0-rc.1" → "3.0.0-rc.1". */
export function shortVersion(version: string): string {
	return version.startsWith('v') ? version.slice(1) : version;
}
