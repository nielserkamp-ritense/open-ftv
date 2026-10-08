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
