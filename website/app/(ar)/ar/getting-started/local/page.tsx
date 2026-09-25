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
  title: "التطوير المحلي",
  description:
    "شغّل واجهة FileShare البرمجية بلغة Go والواجهة الأمامية React محليًا مع إعادة التحميل الفوري.",
};

export default function LocalDevPageAr() {
  return (
    <DocsShell
      locale="ar"
      eyebrow="البدء / محلي"
      title="التطوير المحلي"
      description="عمليتان: واجهة Go البرمجية على المنفذ 3000 وVite على 5173. الوكيل (proxy) مدمج في إعدادات الواجهة الأمامية."
      aside={
        <div className="panel p-5">
          <p className="eyebrow">المنافذ</p>
          <ul className="mt-3 space-y-2 text-sm text-muted">
            <li>
              <code>3000</code> — واجهة HTTP البرمجية
            </li>
            <li>
              <code>5173</code> — واجهة Vite الرسومية
            </li>
            <li>
              <code>50051</code> — gRPC (اختياري)
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
              href="/ar/getting-started/docker/"
              className="block text-accent transition hover:text-paper"
            >
              التالي: Docker ←
            </Link>
          </div>
        </div>
      }
    >
      <ProseH2>المتطلبات السابقة</ProseH2>
      <UL items={["Go 1.25 أو أحدث", "Node.js 20+ وnpm", "Git"]} />

      <Step locale="ar" n="1" title="استنسخ المستودع">
        <CodeBlock
          code={`git clone ${GITHUB_URL}.git
cd simple-file-share`}
        />
      </Step>

      <Step locale="ar" n="2" title="شغّل الواجهة البرمجية والرسومية معًا">
        <P>من جذر المستودع، يشغّل Makefile الخادم وVite في طرفية واحدة:</P>
        <CodeBlock label="الطرفية أ" code={`make dev`} />
        <P>أو شغّلهما منفصلين إن فضّلت طرفية لكل منهما:</P>
        <CodeBlock
          label="الطرفية أ — API"
          code={`cd backend
go mod download
go run ./cmd/server`}
        />
        <CodeBlock
          label="الطرفية ب — الواجهة"
          code={`cd frontend
npm install
npm run dev`}
        />
      </Step>

      <Step locale="ar" n="3" title="الإعداد (اختياري)">
        <P>
          إعدادات التطوير تعمل مباشرة من الصندوق (المصادقة معطّلة، قرص محلي).
          عدّلها حسب حاجتك:
        </P>
        <CodeBlock
          label="env"
          code={`PORT=3000
ROOT_DIR=.file-share-data          # or FILE_SHARE_ROOT
ENABLE_AUTH=false                  # true requires login
ENABLE_SIGNUP=true
ENABLE_TLS=false
APP_ENV=development
ADMIN_USERNAME=admin               # seeded only on first run
ADMIN_PASSWORD=admin`}
        />
        <Callout tone="warn">
          فضّل <code>ADMIN_USERNAME</code>/<code>ADMIN_PASSWORD</code> على{" "}
          <code>USERNAME</code>/<code>PASSWORD</code> القديمة — فقد يعرّف
          الطرفية <code>USERNAME</code> بالفعل.
        </Callout>
      </Step>

      <Step locale="ar" n="4" title="افتح التطبيق">
        <UL
          items={[
            <>
              الواجهة: <code>http://localhost:5173</code>
            </>,
            <>
              حالة الخادم: <code>http://localhost:3000/health</code>
            </>,
            <>
              Swagger: <code>http://localhost:3000/swagger</code>
            </>,
          ]}
        />
        <Callout>
          يحوّل Vite طلبات <code>/api/*</code> إلى الخادم على{" "}
          <code>:3000</code>. إن رأيت «تعذّر الوصول إلى الخادم»، شغّل خادم Go
          أولًا.
        </Callout>
      </Step>

      <ProseH2>أهداف make المفيدة</ProseH2>
      <CodeBlock
        code={`make run          # backend only
make dev          # backend + frontend
make test-backend # go test -race
make lint         # fmt + vet + frontend lint
make frontend-build`}
      />

      <ProseH2>الجوال (اختياري)</ProseH2>
      <CodeBlock
        code={`cd mobile
npm install
echo 'EXPO_PUBLIC_API_URL=http://10.0.2.2:3000' > .env.local  # Android emulator
npx expo start
npx tsc --noEmit && npx expo lint`}
      />
      <P>
        التالي: <Link href="/ar/getting-started/docker/">إعداد Docker</Link> أو{" "}
        <Link href="/ar/getting-started/first-steps/">الخطوات الأولى في التطبيق</Link>
        .
      </P>
    </DocsShell>
  );
}
