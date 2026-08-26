# Remove FK to the Policy ID in the `policy_audit` table

## Executive Summary

Remove FK to the Policy ID in the `policy_audit` table.

## Description

Keeping a Foreign Key reference to a Policy in the `policy_audit` table prevents the Policy resource from being deleted.
The `DELETE /v1/policy/<policy-id>` endpoint of the Manager API failed with a 500 because of this.

The solution is to remove the Foreign Key constraint, since we can not ensure the Policy resource exists anymore.

## Considered Options

1. Use soft delete. But that would make the retrieval more complex keeping the Policy data around has no real usecase.

## Consequences

- Our domain model should ensure a `policy_audit` record contains a valid Policy ID.
