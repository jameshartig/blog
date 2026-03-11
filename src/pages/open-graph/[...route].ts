import { getCollection, type CollectionEntry } from 'astro:content'
import { OGImageRoute } from 'astro-og-canvas'

export const prerender = true

const collectionEntries = await getCollection('posts')

// Map the array of content collection entries to create an object.
// Converts [{ id: 'post.md', data: { title: 'Example', pubDate: Date } }]
// to { 'post.md': { title: 'Example', pubDate: Date } }
const pages = Object.fromEntries(
  collectionEntries.map((entry: CollectionEntry<'posts'>) => [entry.id.replace(/\.(md|mdx)$/, ''), entry.data])
)

export const { getStaticPaths, GET } = await OGImageRoute({
  param: 'route',
  pages,
  getImageOptions: (_path: string, page: CollectionEntry<'posts'>['data']) => ({
    title: page.title,
    description: page.ogDescription || page.description,
    logo: {
      path: 'public/og/og-logo.png',
      size: [72, 72]
    },
    bgGradient: [[255, 255, 255]],
    /*
    bgImage: {
      path: 'public/og/og-bg.png',
      fit: 'fill'
    },
    */
    padding: 64,
    font: {
      title: {
        color: [28, 28, 28],
        size: 68,
        weight: 'Bold'
      },
      description: {
        color: [90, 90, 90],
        size: 40,
        weight: 'Normal'
      }
    }
  })
})
