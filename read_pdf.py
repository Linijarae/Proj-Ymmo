import fitz
doc = fitz.open(r"c:\Users\dupin\Desktop\Proj'ymmo\UF_INFRA + DEV_B2.pdf")
for i, page in enumerate(doc):
    print(f"=== PAGE {i+1} ===")
    print(page.get_text())
