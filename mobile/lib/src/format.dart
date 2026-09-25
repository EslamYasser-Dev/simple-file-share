import 'dart:math' as math;

String formatBytes(int bytes) {
  if (bytes == 0) return '0 B';
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  var i = (math.log(bytes) / math.log(k)).floor();
  if (i > sizes.length - 1) i = sizes.length - 1;
  if (i < 0) i = 0;
  final value = bytes / math.pow(k, i);
  final rounded = double.parse(value.toStringAsFixed(1));
  final trimmed = rounded == rounded.truncateToDouble()
      ? rounded.toInt().toString()
      : rounded.toString();
  return '$trimmed ${sizes[i]}';
}

String formatDate(String iso) {
  if (iso.isEmpty) return '—';
  final d = DateTime.tryParse(iso)?.toLocal();
  if (d == null) return '—';
  final date = '${d.month}/${d.day}/${d.year}';
  final h = d.hour % 12 == 0 ? 12 : d.hour % 12;
  final mm = d.minute.toString().padLeft(2, '0');
  final suffix = d.hour >= 12 ? 'PM' : 'AM';
  return '$date, ${h.toString().padLeft(2, '0')}:$mm $suffix';
}

String baseName(String path) {
  final clean = path.replaceAll(RegExp(r'/+$'), '');
  final i = clean.lastIndexOf('/');
  return i >= 0 ? clean.substring(i + 1) : clean;
}

String parentPath(String path) {
  final clean = path.replaceAll(RegExp(r'/+$'), '');
  final i = clean.lastIndexOf('/');
  return i > 0 ? clean.substring(0, i) : '';
}

String joinPath(String dir, String name) {
  if (dir.isEmpty) return name;
  return '${dir.replaceAll(RegExp(r'/+$'), '')}/$name';
}
