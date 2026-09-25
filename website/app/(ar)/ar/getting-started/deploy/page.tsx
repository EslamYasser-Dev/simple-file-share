import type { Metadata } from "next";
import Link from "next/link";
import {
  Callout,
  CodeBlock,
  DocsShell,
  P,
  ProseH2,
  Step,
  UL,
} from "../../../../../components/site";
import { GITHUB_URL } from "../../../../../lib/site";

export const metadata: Metadata = {
  title: "النشر",
  description:
    "انشر FileShare كحاوية شاملة، أو فصّل بين واجهة GitHub Pages وواجهة API مستضافة.",
};

export default function DeployPageAr() {
  return (
    <DocsShell
      locale="ar"
      eyebrow="البدء / النشر"
      title="النشر"
      description="شكلان: حاوية واحدة تقدّم الواجهتين، أو واجهة ثابتة على GitHub Pages مع خادم API في مكان آخر."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">الأشكال</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <strong className="text-paper">شامل</strong>
              <br />
              مضيف Docker، حاوية واحدة
            </li>
            <li>
              <strong className="text-paper">مقسّم</strong>
              <br />
              واجهة Pages + خادم API منفصل
            </li>
          </ul>
          <div className="mt-5 space-y-2 text-sm">
            <Link
              href="/ar/getting-started/"
              className="block text-accent transition hover:text-paper"
            >
              → كل الأدلة
            </Link>
            <Link
              href="/ar/getting-started/first-steps/"
              className="block text-accent transition hover:text-paper"
            >
              → الخطوات الأولى
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>الشامل (موصى به)</ProseH2>
      <P>
        أي مضيف يدعم Docker (VPS، Fly.io، Koyeb، إلخ): ابنِ الصورة أو اسحبها
        واربط حجمًا بـ <code>/data</code>.
      </P>
      <CodeBlock
        label="docker run"
        code={`docker run -d --name file-share \\
  -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  -e ENABLE_AUTH=true \\
  -e ENABLE_TLS=false \\
  ghcr.io/eslamyasser-dev/simple-file-share:latest`}
      />
      <Callout tone="warn">
        ضع طبقة TLS أمام الخادم (وسيط عكسي أو <code>ENABLE_TLS=true</code>). لا
        ترسل كلمات المرور أبدًا عبر HTTP عادي في الإنتاج.
      </Callout>

      <Step locale="ar" n="1" title="تحقّق">
        <CodeBlock
          code={`curl -fsS https://your-host/health
# open https://your-host/ and sign in`}
        />
      </Step>

      <ProseH2>المقسّم — GitHub Pages + API</ProseH2>
      <P>
        تقدّم Pages نسخة React الثابتة؛ ويعمل خادم Go على مضيف Docker. إعدادات
        CORS متساهلة في الخادم بالفعل.
      </P>

      <Step locale="ar" n="1" title="انشر خادم API">
        <P>
          استخدم التدفق الشامل أعلاه ودوّن المصدر العام، مثل
          https://api.example.com.
        </P>
      </Step>

      <Step locale="ar" n="2" title="انشر الواجهة">
        <UL
          items={[
            <>
              <strong>CI (موصى به):</strong> Settings ← Pages ← Source{" "}
              <em>GitHub Actions</em>؛ ثم أضف متغيّر Actions{" "}
              <code>VITE_API_URL</code> = مصدر خادم API. وخياريًا{" "}
              <code>VITE_BASE_PATH</code>.
            </>,
            <>
              <strong>سكربت:</strong>{" "}
              <code>VITE_API_URL=https://api.example.com ./scripts/deploy-pages.sh</code>{" "}
              — يبني، ويضيف صفحة 404 لـ SPA وملف ‎.nojekyll، ويدفع فرع{" "}
              <code>gh-pages</code> بالقوة.
            </>,
          ]}
        />
        <Callout>
          يفشل CI في الإنتاج إن فُقد <code>VITE_API_URL</code>؛ ومعاينات PR
          تُبنى بدونه. قدّم خادم API عبر HTTPS حتى لا تُرسل بيانات Basic Auth
          أبدًا كنص صريح.
        </Callout>
      </Step>

      <ProseH2>هذا الموقع التعريفي</ProseH2>
      <P>صفحات هذا الموقع في <code>website/</code> تُصدَّر ثابتة:</P>
      <CodeBlock
        code={`cd website
npm install
NEXT_PUBLIC_APP_URL=https://app.example.com \\
NEXT_PUBLIC_GITHUB_URL=${GITHUB_URL} \\
npm run build
# upload website/out/ to any static host`}
      />

      <ProseH2>اختبار تشغيل إنتاجي محلي</ProseH2>
      <CodeBlock
        code={`mkdir -p bin
(cd backend && go build -o ../bin/file-share ./cmd/server)
(cd frontend && npm run build)

APP_ENV=production PORT=8090 ROOT_DIR=./data \\
STATIC_DIR=./frontend/dist \\
JWT_SECRET="$(openssl rand -hex 32)" \\
ADMIN_USERNAME=admin ADMIN_PASSWORD='local-smoke-only' ENABLE_TLS=false \\
./bin/file-share
# UI + API on http://localhost:8090`}
      />

      <ProseH2>قائمة التحقق</ProseH2>
      <UL
        items={[
          "JWT_SECRET عشوائي (32 حرفًا فأكثر) — يرفض الخادم التشغيل في الإنتاج دونه",
          "ADMIN_PASSWORD قوي؛ القيم المعروفة تُرفض عند بذرة أول تشغيل",
          "ENABLE_AUTH=true مع إنهاء TLS",
          "حجم تخزين دائم لـ ROOT_DIR (/data)",
          "MAX_UPLOAD_BYTES إن احتجت حدًا أقصى",
          "ENABLE_SIGNUP=false بعد دعوة الفريق (اختياري)",
          "نسخ احتياطية لحجم البيانات / حاوية S3",
        ]}
      />
      <P>
        جدول المتغيرات الكامل: <code>readme.md</code> في المستودع ← النشر ←
        متغيرات البيئة.
      </P>
    </DocsShell>
  );
}
