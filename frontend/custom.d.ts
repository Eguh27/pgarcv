// Custom type declarations for Next.js
declare module "*.css" {
  const content: { [className: string]: string };
  export default content;
}