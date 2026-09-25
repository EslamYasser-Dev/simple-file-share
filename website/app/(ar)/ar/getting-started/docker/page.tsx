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
  title: "Docker",
  description: "شغّل FileShare عبر Docker Compose أو حاوية واحدة.",
};

export default function DockerPageAr() {
  return (
    <DocsShell
      locale="ar"
      eyebrow="البدء / Docker"
      title="Docker"
      description="صورة شاملة: واجهة React وواجهة Go البرمجية في حاوية واحدة. حجم مسمّى يحتفظ بالملفات تحت ‎/data."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">القيم الافتراضية</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              HTTP <code>22010</code>
            </li>
            <li>
              gRPC <code>50051</code>
            </li>
            <li>
              Volume <code>file-data → /data</code>
            </li>
            <li>
              Compose يتطلب <code>JWT_SECRET</code> +{" "}
              <code>ADMIN_PASSWORD</code>
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
              href="/ar/getting-started/deploy/"
              className="block text-accent transition hover:text-paper"
            >
              التالي: النشر ←
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>الخيار أ — Docker Compose</ProseH2>
      <Step locale="ar" n="1" title="زوّد بيانات الإنتاج السرية">
        <Callout tone="warn">
          <strong>يتطلب</strong> Compose تعيين <code>JWT_SECRET</code> (يرفض
          الخادم التشغيل في الإنتاج دون قيمة عشوائية من 32 حرفًا على الأقل)
          و<code>ADMIN_PASSWORD</code> (لا يوجد تسجيل دخول افتراضي). صدّرهما
          للطرفية أو ضعهما في ملف <code>.env</code> بجوار{" "}
          <code>docker-compose.yml</code> — تقرأهما Docker Compose تلقائيًا.
        </Callout>
        <CodeBlock
          label=".env (بجوار docker-compose.yml)"
          code={`# openssl rand -hex 32
JWT_SECRET=paste-a-random-32-byte-hex-string-here
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-strong-password
ENABLE_SIGNUP=false`}
        />
      </Step>

      <Step locale="ar" n="2" title="استنسخ وشغّل">
        <CodeBlock
          code={`git clone ${GITHUB_URL}.git
cd simple-file-share
docker compose up --build`}
        />
        <P>
          افتح <code>http://localhost:22010</code>. يبقى التخزين في حجم{" "}
          <code>file-data</code>.
        </P>
      </Step>

      <Step locale="ar" n="3" title="أوامر مفيدة">
        <CodeBlock
          code={`docker compose logs -f     # tail logs
docker compose down         # stop (volume kept)
docker compose down -v      # stop and delete volume`}
        />
      </Step>

      <ProseH2>الخيار ب — ابنِ وشغّل</ProseH2>
      <CodeBlock
        code={`docker build -t simple-file-share .

docker run -d --name file-share \\
  -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  simple-file-share`}
      />
      <Callout>
        فضّل حجمًا مسمّى على ربط المجلدات مباشرة. يعمل الخادم بحساب{" "}
        <code>nobody</code> (uid 65534) ويجب أن يملك <code>/data</code> لقفلها
        على <code>0700</code>. بالنسبة لـ{" "}
        <code>{`-v "$PWD/data:/data"`}</code>:
        <br />
        <code>mkdir -p data && sudo chown 65534:65534 data</code>
      </Callout>

      <ProseH2>الخيار ج — الصورة المنشورة</ProseH2>
      <P>
        تُنشر الإصدارات على GitHub Container Registry عند وسم{" "}
        <code>vX.Y.Z</code>:
      </P>
      <CodeBlock
        code={`docker pull ghcr.io/eslamyasser-dev/simple-file-share:latest
docker run -d --name file-share -p 22010:22010 \\
  -v file-share-data:/data \\
  -e APP_ENV=production \\
  -e JWT_SECRET="$(openssl rand -hex 32)" \\
  -e ADMIN_USERNAME=admin \\
  -e ADMIN_PASSWORD='your-strong-password' \\
  ghcr.io/eslamyasser-dev/simple-file-share:latest`}
      />

      <ProseH2>فحص الحالة</ProseH2>
      <CodeBlock
        code={`curl -fsS http://localhost:22010/health
# {"status":"healthy"}`}
      />

      <ProseH2>متغيرات البيئة الأساسية</ProseH2>
      <UL
        items={[
          <>
            <code>APP_ENV=production</code> — يفعّل المصادقة افتراضيًا
          </>,
          <>
            <code>JWT_SECRET</code> — مطلوب في الإنتاج، قيمة عشوائية من 32 حرفًا
            فأكثر
          </>,
          <>
            <code>ADMIN_PASSWORD</code> — مطلوب لبذرة المدير عند أول تشغيل
          </>,
          <>
            <code>PORT</code>، <code>ROOT_DIR</code>، <code>STATIC_DIR</code>
          </>,
          <>
            <code>ENABLE_SIGNUP</code>، <code>ENABLE_TLS</code>،{" "}
            <code>MAX_UPLOAD_BYTES</code>
          </>,
          <>
            <code>STORAGE_BACKEND=local|s3</code> مع بيانات اتصال S3_*
          </>,
        ]}
      />
      <P>
        التالي:{" "}
        <Link href="/ar/getting-started/deploy/">النشر إلى الإنتاج</Link> أو{" "}
        <Link href="/ar/getting-started/first-steps/">الخطوات الأولى</Link>.
      </P>
    </DocsShell>
  );
}
