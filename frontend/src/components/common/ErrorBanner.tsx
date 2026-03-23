import { motion, AnimatePresence } from 'framer-motion'

interface ErrorBannerProps {
  message: string | null
}

export function ErrorBanner({ message }: ErrorBannerProps) {
  return (
    <AnimatePresence>
      {message && (
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
          className="absolute top-16 left-1/2 -translate-x-1/2 bg-white text-black px-4 py-2 rounded font-bold text-sm z-50"
        >
          {message}
        </motion.div>
      )}
    </AnimatePresence>
  )
}
