/**
 * Represents a policy in the system
 */
export interface Policy {
  /** The unique identifier of the policy */
  id: string;

  /** The language of the policy */
  language: string;

  /** The unique identifier of the Register van Verwerkings-Activiteiten (RvVA) */
  rvvaId?: string;

  /** Link to the actual policy */
  url: string;
}
