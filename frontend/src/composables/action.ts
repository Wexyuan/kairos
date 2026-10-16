import type { DialogReactive, InputInst } from 'naive-ui'
import { NInput } from 'naive-ui'

interface ConfirmOptions {
  title: string
  content: string
  action: () => Promise<unknown>
}

interface InputOptions {
  title: string
  initial: string
  action: (value: string) => Promise<unknown>
}

// useAction provides common action utilities.
export function useAction() {
  const message = useMessage()
  const dialog = useDialog()

  // runAction runs a async action.
  async function runAction(label: string, action: () => Promise<unknown>): Promise<boolean> {
    try {
      await action()
      return true
    }
    catch (err) {
      message.error(`${label}失败：${err instanceof Error ? err.message : String(err)}`)
      return false
    }
  }

  // runWithLoading runs a async action with a loading position button.
  async function runWithLoading(instance: DialogReactive, label: string, action: () => Promise<unknown>): Promise<boolean> {
    instance.loading = true
    try {
      return await runAction(label, action)
    }
    finally {
      instance.loading = false
    }
  }

  // confirmAction shows a warning dialog.
  function confirmAction({ title, content, action }: ConfirmOptions): void {
    let instance!: DialogReactive
    const run = () => runWithLoading(instance, title, action)
    instance = dialog.warning({
      title,
      content,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: run,
    })
  }

  // inputAction shows an input dialog.
  function inputAction({ title, initial, action }: InputOptions): void {
    let instance!: DialogReactive
    const value = ref(initial)
    const inputRef = ref<InputInst | null>(null)
    const submit = () => runWithLoading(instance, title, () => action(value.value.trim()))
    const submitAndClose = async () => {
      if (await submit())
        instance.destroy()
    }
    instance = dialog.info({
      title,
      content: () => h(NInput, {
        ref: inputRef,
        value: value.value,
        onUpdateValue: (v: string) => value.value = v,
        onKeydown: (e: KeyboardEvent) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            submitAndClose()
          }
        },
      }),
      positiveText: '确定',
      negativeText: '取消',
      onAfterEnter: () => {
        inputRef.value?.focus()
        inputRef.value?.select()
      },
      onPositiveClick: submit,
    })
  }

  return { runAction, confirmAction, inputAction }
}
