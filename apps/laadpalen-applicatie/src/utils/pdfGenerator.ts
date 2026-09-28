import jsPDF from "jspdf";
import autoTable from "jspdf-autotable";
import {format} from "date-fns";
import {nl} from "date-fns/locale";
import {Aanvraag} from "@/lib/api";

export const generateVergunningPDF = (application: Aanvraag) => {
  const doc = new jsPDF();

  // Header
  doc.setFontSize(20);
  doc.setFont("helvetica", "bold");
  doc.text("GEMEENTE VLIERDAM", 105, 20, {align: "center"});

  doc.setFontSize(12);
  doc.setFont("helvetica", "normal");
  doc.text("Afdeling Vergunningen", 105, 28, {align: "center"});

  // Titel
  doc.setFontSize(16);
  doc.setFont("helvetica", "bold");
  doc.text("VERGUNNING LAADPAAL", 105, 45, {align: "center"});

  // Vergunningsnummer
  doc.setFontSize(10);
  doc.setFont("helvetica", "normal");
  doc.text(`Vergunningsnummer: ${application.id}`, 20, 55);

  // Datum
  doc.text(`Datum verlening: ${format(new Date(), "d MMMM yyyy", {locale: nl})}`, 20, 62);

  // Lijn
  doc.setLineWidth(0.5);
  doc.line(20, 68, 190, 68);

  // Tekst
  doc.setFontSize(11);
  doc.text("Burgemeester en wethouders van de gemeente Vlierdam,", 20, 78);
  doc.text("overwegende dat:", 20, 85);

  doc.setFont("helvetica", "normal");
  doc.text("- de aanvrager voldoet aan de gestelde eisen;", 25, 92);
  doc.text("- de locatie geschikt is voor plaatsing van een laadpaal;", 25, 99);
  doc.text("- er geen bezwaren zijn vanuit omwonenden;", 25, 106);

  doc.setFont("helvetica", "bold");
  doc.text("BESLUITEN:", 20, 118);

  doc.setFont("helvetica", "normal");
  doc.text("Aan onderstaande aanvrager vergunning te verlenen voor het plaatsen", 20, 126);
  doc.text("van een laadpaal op eigen terrein:", 20, 133);

  // Gegevens tabel
  autoTable(doc, {
    startY: 140,
    head: [["Gegeven", "Waarde"]],
    body: [
      ["Adres", `${application.postcode} ${application.huisnummer}`],
      ["Aanvraagdatum", format(application["status-datum"], "d MMMM yyyy", {locale: nl})],
      ["Status", "GOEDGEKEURD"],
    ],
    theme: "grid",
    headStyles: {fillColor: [41, 128, 185]},
    margin: {left: 20, right: 20},
  });

  // Voorwaarden
  const finalY = (doc as jsPDF & { lastAutoTable: { finalY: number } }).lastAutoTable.finalY + 10;
  doc.setFontSize(11);
  doc.setFont("helvetica", "bold");
  doc.text("Voorwaarden:", 20, finalY);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.text("1. De laadpaal dient te voldoen aan de geldende technische eisen.", 20, finalY + 7);
  doc.text("2. De installatie dient uitgevoerd te worden door een erkend installateur.", 20, finalY + 14);
  doc.text("3. De vergunning is perceel gebonden.", 20, finalY + 21);
  doc.text("4. Bij verkoop van de woning is een nieuwe aanvraag niet nodig.", 20, finalY + 28);

  // Handtekening sectie
  const signatureY = finalY + 45;
  doc.text("Namens het college van burgemeester en wethouders,", 20, signatureY);

  // Handtekening (text-based)
  doc.setFont("times", "italic");
  doc.setFontSize(16);
  doc.text("M. de Vries", 20, signatureY + 15);

  doc.setFont("helvetica", "normal");
  doc.setFontSize(10);
  doc.text("Mw. M. de Vries", 20, signatureY + 22);
  doc.text("Hoofd afdeling Vergunningen", 20, signatureY + 28);

  // Footer
  doc.setFontSize(8);
  doc.setTextColor(128, 128, 128);
  doc.text("Gemeente Vlierdam | Postbus 1234 | 1111 AA Vlierdam | www.vlierdam.nl", 105, 285, {align: "center"});

  // Save
  doc.save(`Vergunning-Laadpaal-${application.id}.pdf`);
};