Feature: Scenario 1, vergunning aanvragen
  The testcases of scenario 1 of the Laadpalen demo:
  https://vng-realisatie.github.io/ftv/toepassen/demonstratie/

  Scenario Outline: Testcase <nr>: <gebruiker> may not submit an aanvraag
    Given I open the Laadpalen UI as "<gebruiker>"
    When I submit an aanvraag for postcode "<postcode>" and huisnummer "<huisnummer>"
    Then I see "<melding>"
    And the Logboek of "Gemeente Vlierdam" shows my aanvraag as Geweigerd because of "<reden>"

    Examples:
      | nr | gebruiker | postcode | huisnummer | melding                                            | reden              |
      | 1  | Rick      | 1111AA   | 1          | Rick niet geautoriseerd vanwege afdeling           | afdeling           |
      | 2  | Morty     | 1111AA   | 1          | Morty niet geautoriseerd vanwege opleiding         | opleiding          |
      | 3  | Beth      | 1111AA   | 1          | Beth niet geautoriseerd vanwege verlopen opleiding | verlopen opleiding |

  Scenario Outline: Testcase <nr>: the BRP has no ingezetene at <postcode> <huisnummer>
    Given I open the Laadpalen UI as "Jerry"
    When I submit an aanvraag for postcode "<postcode>" and huisnummer "<huisnummer>"
    And I run the BRP controle
    Then the aanvraag shows "0 bewoners gevonden"
    And the Logboek of "Gemeente Vlierdam" shows the BRP controle as Toegestaan for doelbinding "identificatie-en-ingezetenschap" in 1 decision
    And the Logboek of "RvIG" shows the BRP controle as Toegestaan for doelbinding "identificatie-en-ingezetenschap" in 1 decision

    Examples:
      | nr | postcode | huisnummer |
      | 4  | 1111FF   | 11         |
      | 5  | 2222JJ   | 101        |

  Scenario Outline: Testcase <nr>: the aanvraag can only be rejected because <reden>
    Given I open the Laadpalen UI as "Jerry"
    When I submit an aanvraag for postcode "<postcode>" and huisnummer "<huisnummer>"
    And I run the BRP controle
    And I run the RDW controle
    And I run the Laadpaal controle
    Then the aanvraag shows "<melding>"
    And the aanvraag can only be rejected
    And the Logboek of "Gemeente Vlierdam" shows the RDW controle as Toegestaan for doelbinding "laadinfra-verificatie" in 2 decisions
    And the Logboek of "RDW" shows the RDW controle as Toegestaan for doelbinding "laadinfra-verificatie" in 2 decisions

    Examples:
      | nr | postcode | huisnummer | reden                          | melding                                                   |
      | 6  | 1111DD   | 4          | the BRV has no voertuig        | Geen gegevens beschikbaar                                 |
      | 7  | 1111CC   | 3          | the voertuig is not electric   | Geen elektrische voertuigen (uitstootklasse E) gevonden   |
      | 9  | 1111AA   | 1          | the address has a laadpaal     | Er is al een laadpaalvergunning uitgegeven voor dit adres |

  Scenario: Testcase 8: the RDW withholds a diplomatic voertuig
    Given I open the Laadpalen UI as "Jerry"
    When I submit an aanvraag for postcode "1111EE" and huisnummer "5"
    And I run the BRP controle
    And I run the RDW controle
    Then the aanvraag shows "RDW verstrekt geen gegevens over diplomatieke voertuigen"
    And the Logboek of "Gemeente Vlierdam" shows the RDW controle as Toegestaan for doelbinding "laadinfra-verificatie" in 2 decisions
    And the Logboek of "RDW" shows the RDW controle as Toegestaan in 1 decision and as Geweigerd in 1 decision because of "RDW verstrekt geen gegevens over diplomatieke voertuigen"

  Scenario: Testcase 10: the aanvraag can be approved
    Given I open the Laadpalen UI as "Jerry"
    When I submit an aanvraag for postcode "1111BB" and huisnummer "2"
    And I run the BRP controle
    And I run the RDW controle
    And I run the Laadpaal controle
    Then the aanvraag can be approved
    And the Logboek of "Gemeente Vlierdam" shows the RDW controle as Toegestaan for doelbinding "laadinfra-verificatie" in 2 decisions
    And the Logboek of "RDW" shows the RDW controle as Toegestaan for doelbinding "laadinfra-verificatie" in 2 decisions
