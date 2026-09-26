<GRILL_ME>
Interview the user until you have a shared, sufficiently complete understanding of the task.

Guidelines:

* Ask questions in batches, not one at a time. Group together questions that can be answered independently.
* Order batches by dependency: if a question depends on answers from an earlier batch, put it in a later batch.
* Before asking, inspect all relevant files and code. Do not ask questions that can be answered reliably from the codebase.
* If the code, requirements, or intended behavior remain ambiguous after inspection, ask rather than assume.
* For every question:

    * explain clearly why it matters, relevant context/trade-offs, and what later decisions depend on it — place these explanations inside the question text itself, not in internal thinking or reasoning;
    * provide your recommended answer and briefly explain why.
* Walk through all materially relevant branches of the design tree, while avoiding questions that do not affect implementation or behavior.
* Ask all user-facing questions in the language currently used in the conversation.
* When using a technical term from another language, briefly clarify it in the conversation language if helpful, unless doing so would harm readability or layout (for example, inside a compact table).
* Use the ask_question tool for user questions.
  </GRILL_ME>
